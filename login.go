package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	g "github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

var (
	conf = &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_ID"),
		ClientSecret: os.Getenv("GITHUB_SECRET"),
		Scopes:       []string{"user:email"},
		RedirectURL:  os.Getenv("GITHUB_REDIRECT"),
		Endpoint:     github.Endpoint,
	}
	playground_secret = os.Getenv("PLAYGROUND_SECRET")
)

type User struct {
	Email    string
	Gravatar string
}

func (u *User) New(email string) {
	u.Email = email
	hasher := md5.New()
	hasher.Write([]byte(email))
	u.Gravatar = "http://www.gravatar.com/avatar/" + hex.EncodeToString(hasher.Sum(nil)) + "?s=256"
}

// Examines JWT token found in cookie to see if user is authenticated
func authenticated(req *http.Request) (bool, *User) {
	cookie, err := req.Cookie("token")
	if err != nil {
		return false, nil
	}

	if cookie != nil {
		tokenString := cookie.Value

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(playground_secret), nil
		})

		if err != nil {
			return false, nil
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			if seconds, ok := claims["exp"].(float64); ok {
				expiration := time.Unix(int64(seconds), 0)
				if expiration.Sub(time.Now()) > 0 {
					if email, ok := claims["email"].(string); ok {
						u := &User{}
						u.New(email)
						return true, u
					}
				}
			}
		}
	}

	return false, nil
}

func encrypt(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, aes.BlockSize+len(text))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(text))
	return ciphertext, nil
}

func decrypt(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(text) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}
	iv := text[:aes.BlockSize]
	text = text[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(text, text)

	if err != nil {
		return nil, err
	}
	return text, nil
}

func Login(w http.ResponseWriter, r *http.Request) {
	timestamp, err := time.Now().MarshalText()
	if err != nil {
		// Redirect to error page
		return
	}

	cipher, err := encrypt([]byte(playground_secret), timestamp)
	if err != nil {
		// Redirect to error page
		return
	}

	state := base64.StdEncoding.EncodeToString([]byte(cipher))
	state = url.QueryEscape(state)
	url := conf.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{Name: "token", Value: "", HttpOnly: true}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func Callback(w http.ResponseWriter, r *http.Request) {
	// No matter what happens, redirect back to home page. If login succesful it will load the full application, otherwise just the playground
	defer http.Redirect(w, r, "/", http.StatusTemporaryRedirect)

	unescaped, err := url.QueryUnescape(r.FormValue("state"))
	if err != nil {
		return
	}

	decoded, err := base64.StdEncoding.DecodeString(unescaped)
	if err != nil {
		return
	}

	state, err := decrypt([]byte(playground_secret), decoded)
	if err != nil {
		return
	}

	dt := time.Time{}
	dt.UnmarshalText(state)

	elapsed := time.Now().Sub(dt)
	if elapsed.Minutes() > 1 {
		return
	}

	code := r.FormValue("code")
	token, err := conf.Exchange(oauth2.NoContext, code)
	if err != nil {
		return
	}

	jsToken, err := tokenToJSON(token)
	if err != nil {
		return
	}

	oauthClient := conf.Client(oauth2.NoContext, token)
	client := g.NewClient(oauthClient)
	opt := &g.ListOptions{}
	opt.Page = 1
	opt.PerPage = 30
	emails, _, err := client.Users.ListEmails(opt) // Don't use User.Get, e-mail might be private

	if err != nil {
		return
	}

	if len(emails) == 0 {
		return
	}

	email := *emails[0].Email // Go with first e-mail in list

	t := &TokenData{}
	t.New(email, jsToken)

	err = StoreToken(t) // Store token in DynamoDB
	if err != nil {
		return
	}

	expiration := time.Now().Add(30 * 24 * time.Hour)

	// Generate JWT
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": email,
		"nbf":   time.Now().Unix(),
		"exp":   expiration.Unix(),
		"iss":   "dodolabs.io",
	})

	tokenString, err := jwtToken.SignedString([]byte(playground_secret))
	if err != nil {
		return
	}

	// Put in cookie
	cookie := http.Cookie{Name: "token", Value: tokenString, Expires: expiration, HttpOnly: true}
	http.SetCookie(w, &cookie)
}

func tokenToJSON(token *oauth2.Token) (string, error) {
	if d, err := json.Marshal(token); err != nil {
		return "", err
	} else {
		return string(d), nil
	}
}

func tokenFromJSON(jsonStr string) (*oauth2.Token, error) {
	var token oauth2.Token
	if err := json.Unmarshal([]byte(jsonStr), &token); err != nil {
		return nil, err
	}
	return &token, nil
}

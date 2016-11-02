package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"hash/crc32"
	"strconv"
)

type TokenData struct {
	Email string
	Token string
}

type Binary struct {
	Id      string
	Source  string
	Fram    []byte
	Results string
	Version string
}

type Project struct {
	Title  string `json:"title"`
	Source string `json:"source,omitempty"`
}

func (t *TokenData) New(email, token string) {
	t.Email = email
	t.Token = token
}

func StoreToken(t *TokenData) error {
	params := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"Email": {
				S: aws.String(t.Email),
			},
			"Token": {
				S: aws.String(t.Token),
			},
		},
		TableName: aws.String("Tokens"),
	}

	_, err := svc.PutItem(params)

	return err
}

func GetToken(email string) (*TokenData, error) {
	params := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"Email": {
				S: aws.String(email),
			},
		},
		TableName:      aws.String("Tokens"),
		ConsistentRead: aws.Bool(true),
	}

	resp, err := svc.GetItem(params)

	if err != nil {
		return nil, err
	}

	item := resp.Item
	if item == nil {
		return nil, nil
	}

	t := &TokenData{
		Email: email,
		Token: aws.StringValue(item["Token"].S),
	}

	return t, nil
}

func (b *Binary) New(id string, source string, fram []byte, results string, version string) {
	b.Id = id
	b.Source = source
	if fram != nil {
		b.Fram = fram
	} else {
		b.Fram = []byte{0}
	}
	b.Results = results
	b.Version = version
}

func StoreBinary(b *Binary) error {
	params := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"Id": {
				S: aws.String(b.Id),
			},
			"Source": {
				S: aws.String(b.Source),
			},
			"Fram": {
				B: b.Fram,
			},
			"Results": {
				S: aws.String(b.Results),
			},
			"Version": {
				S: aws.String(b.Version),
			},
		},
		TableName: aws.String("Binaries"),
	}

	_, err := svc.PutItem(params)

	return err
}

func GetBinary(id string) (*Binary, error) {
	params := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"Id": {
				S: aws.String(id),
			},
		},
		TableName:      aws.String("Binaries"),
		ConsistentRead: aws.Bool(true),
	}

	resp, err := svc.GetItem(params)

	if err != nil {
		return nil, err
	}

	item := resp.Item
	if item == nil {
		return nil, nil
	}

	b := &Binary{
		Id:      id,
		Source:  aws.StringValue(item["Source"].S),
		Fram:    item["Fram"].B,
		Results: aws.StringValue(item["Results"].S),
		Version: aws.StringValue(item["Version"].S),
	}

	return b, nil
}

func GenerateCRC(source string) string {
	r := crc32.ChecksumIEEE([]byte(source))
	return strconv.FormatUint(uint64(r), 16)
}

func GetProjects(email string) ([]string, error) {
	params := &dynamodb.QueryInput{
		KeyConditionExpression: aws.String("Email = :email"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":email": {
				S: aws.String(email),
			},
		},
		TableName: aws.String("Projects"),
	}

	resp, err := svc.Query(params)
	if err != nil {
		return nil, err
	}
	count := aws.Int64Value(resp.Count)

	ret := make([]string, count)
	for i, item := range resp.Items {
		ret[i] = aws.StringValue(item["Title"].S)
	}

	return ret, nil
}

func GetProject(email, title string) (*Project, error) {
	params := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"Email": {
				S: aws.String(email),
			},
			"Title": {
				S: aws.String(title),
			},
		},
		TableName:      aws.String("Projects"),
		ConsistentRead: aws.Bool(true),
	}

	resp, err := svc.GetItem(params)

	if err != nil {
		return nil, err
	}

	item := resp.Item
	if item == nil {
		return nil, nil
	}

	p := &Project{
		Title:  title,
		Source: aws.StringValue(item["Source"].S),
	}

	return p, nil
}

func StoreProject(email string, p *Project) error {
	params := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"Email": {
				S: aws.String(email),
			},
			"Title": {
				S: aws.String(p.Title),
			},
			"Source": {
				S: aws.String(p.Source),
			},
		},
		TableName: aws.String("Projects"),
	}

	_, err := svc.PutItem(params)

	return err
}

func CreateProject(email string, p *Project) error {
	params := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"Email": {
				S: aws.String(email),
			},
			"Title": {
				S: aws.String(p.Title),
			},
			"Source": {
				S: aws.String(p.Source),
			},
		},
		TableName:           aws.String("Projects"),
		ConditionExpression: aws.String("attribute_not_exists(Title)"),
	}

	_, err := svc.PutItem(params)

	return err
}

func DeleteProject(email, title string) error {
	params := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"Email": {
				S: aws.String(email),
			},
			"Title": {
				S: aws.String(title),
			},
		},
		TableName: aws.String("Projects"),
	}

	_, err := svc.DeleteItem(params)

	return err
}

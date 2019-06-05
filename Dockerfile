FROM golang:1.12

RUN mkdir -p /go/src
ADD . /go/src
WORKDIR /go/src

RUN go get -u github.com/gorilla/mux
RUN go get -u github.com/gopherjs/jquery
RUN go get -u github.com/peternoyes/dodo-sim
RUN go get -u github.com/russross/blackfriday

RUN go get -u github.com/dgrijalva/jwt-go
RUN go get -u github.com/google/go-github/github
RUN go get -u golang.org/x/oauth2
RUN go get -u golang.org/x/oauth2/github

RUN go get -u github.com/flimzy/jsblob

RUN go get -u -d github.com/gopherjs/gopherpen/...
RUN go get -u -d -tags=dev github.com/gopherjs/gopherpen/...
RUN go get -u -d -tags=generate github.com/gopherjs/gopherpen/...

RUN go get -u github.com/aws/aws-sdk-go/...

RUN go generate
RUN go build -o dodo-playground

CMD ["./dodo-playground"]

EXPOSE 3000

RUN git clone https://github.com/cc65/cc65 /home/cc65

RUN cd /home/cc65 \
	&& make
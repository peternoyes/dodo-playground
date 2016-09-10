FROM golang:1.7

RUN mkdir -p /go/src
ADD . /go/src
WORKDIR /go/src

RUN go get -u github.com/gorilla/mux
RUN go get -u github.com/gopherjs/jquery
RUN go get -u github.com/peternoyes/dodo-sim
RUN go get -u github.com/peternoyes/httpgzip
RUN go get -u github.com/russross/blackfriday

RUN go get -u -d github.com/gopherjs/gopherpen/...
RUN go get -u -d -tags=dev github.com/gopherjs/gopherpen/...
RUN go get -u -d -tags=generate github.com/gopherjs/gopherpen/...

RUN go build -tags=dev -o dodo-playground

CMD ["./dodo-playground"]

EXPOSE 3000

RUN git clone https://github.com/cc65/cc65 /home/cc65

RUN cd /home/cc65 \
	&& make
# For Go 1.7
FROM golang:1.7-onbuild
EXPOSE 3000

RUN git clone https://github.com/cc65/cc65 /home/cc65

RUN cd /home/cc65 \
	&& make

# dodo-playground

Dodo-playground is a website that allows users to write and run [Dodo](http://www.dodolabs.io) games. The entire stack of the website is written in Go!

Dodo is a 6502 homebrew Game System. The original 6502 emulator and Dodo simulator was written in Go as a console application. The intent was that the web version would have the simulation run server side and stream the graphics down to the client. In theory this works, but it doesn't scale. 

The playground project was put on hold because porting the simulator to JavaScript just didn't sound fun. Thankfully, I stumbled across [GopherJS](http://www.gopherjs.org/) which transpiles Go to JavaScript for front end web development. 

# Technology

- Built from the [Gopherpen](https://github.com/gopherjs/gopherpen) template for GopherJS
- Ace Editor
- Bootstrap
- cc65 6502 C compiler on the backend
- Built to be hosted in the AWS Elastic Beanstalk using Docker
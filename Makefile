.PHONY: all
all: rbg2p zip

rbg2p_lin: 
	GOOS=linux GOARCH=amd64 go build -o g2p cmd/g2p/*go
	GOOS=linux GOARCH=amd64 go build -o server cmd/server/*go

rbg2p_win:
	GOOS=windows GOARCH=amd64 go build -o g2p.exe cmd/g2p/*go
	GOOS=windows GOARCH=amd64 go build -o server.exe cmd/server/*go

rbg2p: clean rbg2p_lin rbg2p_win

zip: clean rbg2p
	zip -q -r rbg2p.zip g2p.exe g2p server.exe server README.md

clean:
	rm -rf g2p.exe g2p server.exe server


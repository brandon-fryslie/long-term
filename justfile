build:
    go build -o long-term .

install: build
    mkdir -p ~/.local/bin
    cp long-term ~/.local/bin/

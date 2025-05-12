FROM golang:1.24.3-alpine AS builder

WORKDIR /usr/src/app

# Primero copiar solo los archivos de dependencias
COPY go.mod go.sum ./
RUN go mod download

# Luego copiar el resto y compilar
COPY . .
RUN go build -v -o /usr/local/bin/app .

EXPOSE 8080
CMD ["app"]

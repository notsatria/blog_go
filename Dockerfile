# Stage 1: Build (Kita pakai image Golang resmi untuk compile)
FROM golang:1.25.4 AS builder

WORKDIR /app

# Copy file dependency dulu biar cache-nya awet
COPY go.mod go.sum ./
RUN go mod download

# Copy semua codingan
COPY . .

# Build aplikasi jadi binary file namanya "main"
RUN go build -o main .

# Stage 2: Run (Kita pakai image Alpine yang kecil banget untuk jalanin apps)
FROM alpine:latest

WORKDIR /root/

# Copy hasil build dari Stage 1
COPY --from=builder /app/main .
COPY --from=builder /app/index.html . 

# Expose port 8080
EXPOSE 8080

# Jalankan aplikasinya
CMD ["./main"]
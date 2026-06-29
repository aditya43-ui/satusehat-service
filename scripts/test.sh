#!/bin/bash

echo "🧪 Starting gRPC Test Suite..."
echo "================================"

# Cek apakah server sudah berjalan
if ! nc -z localhost 8090 2>/dev/null; then
    echo "⚠️  gRPC Server not running on port 8090"
    echo "🚀 Starting server..."
    
    # Jalankan server di background
    go run cmd/api/main.go &
    SERVER_PID=$!
    
    # Tunggu server siap
    echo "⏳ Waiting for server to start..."
    sleep 5
    
    # Cek lagi
    if ! nc -z localhost 8090 2>/dev/null; then
        echo "❌ Failed to start server"
        exit 1
    fi
    
    echo "✅ Server started successfully"
else
    echo "✅ Server already running"
fi

echo ""
echo "1️⃣ Testing gRPC Connection..."
go test -run TestGRPCConnection ./internal/infrastructure/transport/grpc/proto/test/ -v

echo ""
echo "2️⃣ Testing Create Person..."
go test -run TestCreatePerson ./internal/infrastructure/transport/grpc/proto/test/ -v

echo ""
echo "3️⃣ Testing List Persons..."
go test -run TestListPersons ./internal/infrastructure/transport/grpc/proto/test/ -v

echo ""
echo "4️⃣ Testing Get Person..."
go test -run TestGetPerson ./internal/infrastructure/transport/grpc/proto/test/ -v

echo ""
echo "5️⃣ Testing Update Person..."
go test -run TestUpdatePerson ./internal/infrastructure/transport/grpc/proto/test/ -v

echo ""
echo "6️⃣ Testing Delete Person..."
go test -run TestDeletePerson ./internal/infrastructure/transport/grpc/proto/test/ -v

echo ""
echo "7️⃣ Testing Performance..."
go test -run TestGRPCPerformance ./internal/infrastructure/transport/grpc/proto/test/ -v

echo ""
echo "🏁 Test completed!"

# Jika kita yang start server, matikan lagi
if [ ! -z "$SERVER_PID" ]; then
    echo "🛑 Stopping server..."
    kill $SERVER_PID
fi
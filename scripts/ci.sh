cd cli

echo "go vet..."
go vet ./...
echo ""

echo "go build..."
go build ./...
echo ""

echo "go test..."
go test ./...
echo ""

echo "go build spec..."
go build -o ../build/spec ./spec
echo ""

cd ..
echo "spec index..."
build/spec index
echo ""

echo "spec check..."
build/spec check --strict
echo ""
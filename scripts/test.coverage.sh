# 1) cd to root
cd backend

# 2) test cache clean
go clean -testcache

mkdir ../tests/

# 3) run test and generate coverage file
go test ./... -coverprofile=../tests/coverage.out

# 4) show coverage in terminal
go tool cover -func=../tests/coverage.out | tail -n 1

# 5) HTML coverage report
go tool cover -html=../tests/coverage.out -o ../tests/coverage.html

# 6) open in browser
open ../tests/coverage.html
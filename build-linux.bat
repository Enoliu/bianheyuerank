@echo off
echo Building frontend...
cd frontend
call npm run build
cd ..

echo Copying dist to backend...
if exist backend\dist rmdir /s /q backend\dist
xcopy /E /I /Q frontend\dist backend\dist

echo Building backend for Linux...
cd backend
set GOOS=linux
set GOARCH=amd64
go build -o ../server-linux
cd ..

echo Done! Output: server-linux

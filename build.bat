@echo off
echo Building frontend...
cd frontend
call npm run build
cd ..

echo Copying dist to backend...
if exist backend\dist rmdir /s /q backend\dist
xcopy /E /I /Q frontend\dist backend\dist

echo Building backend...
cd backend
go build -o ..\server.exe
cd ..

echo Done! Run server.exe to start on port 8082.

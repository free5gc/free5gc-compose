# free5GC Web Console

### Install Node.js
```bash
sudo apt remove nodejs -y
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt update
sudo apt install nodejs -y
node -v # check that version is 20.x
sudo corepack enable
```

### Build the Server

To be able to run free5gc's webconsole server, consider building its source through the following steps:

```bash
# (In directory: ~/free5gc/webconsole)
cd frontend
yarn install
yarn build
rm -rf ../public
cp -R build ../public
```

### Run the Server

To run free5gc's webconsole server, use:

```bash
# (In directory: ~/free5gc/webconsole)
go run server.go
```

### Connect to WebConsole

Enter `<WebConsole server's IP>:5000` in an internet browser URL bar

Then use the credentials below:
- Username: admin
- Password: free5gc

## Run the Frontend Dev Web Server
Run the frontend development server with file watcher
```bash
cd frontend/
yarn start
```

To specify backend server api url
```bash
cd frontend/
REACT_APP_HTTP_API_URL=http://127.0.0.1:5000/api PORT=3000 yarn start
```

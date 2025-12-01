# free5GC Web Console Frontend

Note that this tutorial is for frontend development ONLY, not the whole web console!

## Environment Setup for Frontend Development

Install yarn
```bash
curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | sudo apt-key add -
echo "deb https://dl.yarnpkg.com/debian/ stable main" | sudo tee /etc/apt/sources.list.d/yarn.list
sudo apt update && sudo apt install yarn
```

Install required packages
```bash
yarn install
```

Note that the default api server base request url is defined in:
`webconsole/frontend/src/config/config.js`

## Run the Frontend Dev Web Server
Run the frontend development server with file watcher
```bash
yarn start
```

To specify backend server api url
```
REACT_APP_HTTP_API_URL=http://127.0.0.1:5000/api PORT=3000 yarn start
```

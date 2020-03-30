**BST API**

[Click Me for Endpoints](endpoints.md)

WIP api leveraging Auth0 or another identity provider.

Build for *nix with:

```
env GOOS=linux GOARCH=amd64 go build
``` 

Or for your own OS with:
```
go build
```

Run with:

```
./bst_web \
    -issuer="https://myissuer..com/" \
    -audience="myaudience" \
    -host="api.bst.com" \
    -port="443" \
    -apibase="/" \
    -dbuser="dbuser" \
    -dbpass="dbpass" \
    -dbname="dbname" \
    -dbhost="1.2.3.4" \
    -dbmigrate=false
```

Setting dbmigrate to `true` will setup/migrate tables.

---

**Setting up on vm**

- Bring up VM
- Create pub/priv key pair for instance. Add to vm for ssh communication.
- Create cert_cache_api directory in planned executable directory.
- Create service bst_api.service
- sudo chmod 664 /etc/systemd/system/bst_api.service
- Edit /etc/security/limits.conf
    - bst soft nofile 4096
    - bst hard nofile 8192
- Edit /etc/systemd/user.conf
    - DefaultLimitNOFILE=4096
- Edit /etc/systemd/system.conf
    - DefaultLimitNOFILE=16384
- Transfer binary to planned executable directory
- sudo systemctl daemon-reload
- sudo systemctl restart bst_api

**Making API Calls**

Endpoints may currently be either public or protected. There is not yet any concept of scope, however this will be added in the future.

Protected endpoints require the `Authorization` header with the value `Bearer mytoken`. 

At some point, a Postman collection will be added to fast-track usage of the API.

---

*TODO*
 - Improve this readme
 - Add endpoints
 - ???
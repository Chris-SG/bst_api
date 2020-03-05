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
    -issuer="https://auth-issuer.com/" \
    -audience="myaudience" \
    -host="api.mysite.com" \
    -port="8443" \
    -apibase="/" \
    -dbuser="dbuser" \
    -dbpass="dbpassword" \
    -dbname="dbname" \
    -dbhost="1.2.3.4"
```

---

**Making API Calls**

Endpoints may currently be either public or protected. There is not yet any concept of scope, however this will be added in the future.

Protected endpoints require the `Authorization` header with the value `Bearer mytoken`. 

At some point, a Postman collection will be added to fast-track usage of the API.

---

*TODO*
 - Improve this readme
 - Add endpoints
 - ???
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
 
 --- 
 
 *Error Messages*
 _jwt_err_: could not extract auth0 name
 _no_user_: no eagate user exists
 _no_cookie_: no cookie found for user
 _bad_cookie_: user cookie has expired
 
 _ddr_songid_fail_: could not get ddr song ids
 _ddr_songid_db_fail_: could not get ddr song ids from database
 _ddr_songdata_fail_: could not get song data for user
 _ddr_songdiff_fail": could not get diff data from eagate
 _ddr_addsong_fail_: could not add songs to database
 _ddr_adddiff_fail_: could not add diff data to database
 _ddr_retdiff_fail_: could not retrieve diff data from database
 _ddr_pi_fail_: could not get player info from eagate
 _ddr_addpi_fail_: could not add player info to database
 _ddr_retpi_fail_: could not retrieve player info from database
 _ddr_addpc_fail_: could not add playcount data to database
 _ddr_retpc_fail_: could not retrieve playcount data from database
 _ddr_songstat_fail_: could not get song stats from eagate
 _ddr_addsongstat_fail_: could not add song stats to database
 _ddr_recent_fail_: could not get recent scores from eagate
 _ddr_addscore_fail_: could not add recent scores to database
 _ddr_wd_fail_: could not load workout data from eagate
 _ddr_addwd_fail_: could not add workout data to database
 _ddr_retwd_fail_: could not retrieve workout data from database
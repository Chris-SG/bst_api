# Endpoints

## root endpoints: `/`

### `/status` http.GET ✅
Current API status.

*headers*
```json

```
*payload*
```json
{
}
```
*response*
```json
{
  "api": "ok",
  "gate": "ok",
  "db": "ok"
}
```


## DDR endpoints: `/ddr`

### `/profile/update` http.PATCH ✅
Update user profile with latest statistics and scores.

*headers*
```
    "Authorization": "Bearer {{bearer_token}}"
```
*payload*
```json
{
}
```
*response*
```json
{
  "status": "ok"
}
```

### `/profile/refresh` http.PATCH ✅
Re-process statistics for all difficulties.

*headers*
```
    "Authorization": "Bearer {{bearer_token}}"
```
*payload*
```json
{
}
```
*response*
```json
{
  "status": "ok"
}
```

### `/songs` http.GET ✅
List of songs currently in the database.

*headers*
```

```
*payload*
```json
{
}
```
*response*
```json
[
  {
    "Id":"1a2b3c4d5e6f",
    "Name":"My First Song",
    "Artist":"Bemani Sound Team"
  },
  {
    "Id":"a1b2c3d4e5f6",
    "Name":"My Second Song",
    "Artist":"Bemani Sound Team"
  },
  ...
]
```

### `/songs` http.PATCH ✅
Update songs in database.

*headers*
```
    "Authorization": "Bearer {{bearer_token}}"
```
*payload*
```json
{
}
```
*response*
```json
{
  "status": "ok",
  "message": "added 100 new songs (800 new difficulties)"
}
```

### `/songs/jackets` http.GET ✅
List of jackets for provided songs.

*headers*
```json
```
*payload*
```json
{
  "ids":
  [
    "1a2b3c4d5e6f",
    "a1b2c3d4e5f6",
    ...
  ]
}
```
*response*
```json
[
  {
    "id": "1a2b3c4d5e6f",
    "jacket": "base64encoded="
  },
  {
    "id": "a1b2c3d4e5f6",
    "jacket": "base64encoded="
  },
  ...
]
```

### `/songs/{id: song_id}` http.GET ✅
Get details for provided song id.

*headers*
```json
```
*payload*
```json
{
}
```
*response*
```json
{
  "id":"1a2b3c4d5e6f",
  "name":"A song name",
  "artist":"A song artist",
  "image":"abase64encodedimage=",
  "difficulties":
  [
    {
      "mode": "SINGLE",
      "difficulty": "BEGINNER",
      "difficultyvalue": 1
    },
    {
      "mode": "SINGLE",
      "difficulty": "BASIC",
      "difficultyvalue": 3
    },
    ...
  ]
}
```

### `/songs/scores` http.GET
List of users top scores.

*headers*
```json
    "Authorization": "Bearer {{bearer_token}}"
```
*payload*
```json
{
  "ids": [
    "{{song_id}}",
    "{{song_id}}",
    ...
  ], OPTIONAL
  "order_by": "{{field_name}}" OPTIONAL
}
```
*response*
```json
[
  {
    "song_id": "{{song_id}}",
    "mode": "{{mode}}",
    "difficulty": "{{difficulty}}",
    "best_score": {{highscore}},
    "lamp": "{{lamp}}",
    "rank": "{{rank}}",
    "playcount": {{playcount}},
    "clearcount": {{clearcount}},
    "maxcombo": {{maxcombo}},
    "last_played": {{last_played}}
  },
  {
    "song_id": "{{song_id}}",
    "mode": "{{mode}}",
    "difficulty": "{{difficulty}}",
    "best_score": {{highscore}},
    "lamp": "{{lamp}}",
    "rank": "{{rank}}",
    "playcount": {{playcount}},
    "clearcount": {{clearcount}},
    "maxcombo": {{maxcombo}},
    "last_played": {{last_played}}
  },
  ...
]
```

### `/songs/scores/{id: song_id}` http.GET
List of users scores for given song.

*headers*
```json
    "Authorization": "Bearer {{bearer_token}}"
```
*payload*
```json
{
  "order_by": "{{field_name}}" OPTIONAL
}
```
*response*
```json
{
  "song_id": "{{song_id}}",
  "top_scores":
  [
    {
      "mode": "{{mode}}",
      "difficulty": "{{difficulty}}",
      "best_score": {{highscore}},
      "lamp": "{{lamp}}",
      "rank": "{{rank}}",
      "playcount": {{playcount}},
      "clearcount": {{clearcount}},
      "maxcombo": {{maxcombo}},
      "last_played": {{last_played}}
    },
    {
      "mode": "{{mode}}",
      "difficulty": "{{difficulty}}",
      "best_score": {{highscore}},
      "lamp": "{{lamp}}",
      "rank": "{{rank}}",
      "playcount": {{playcount}},
      "clearcount": {{clearcount}},
      "maxcombo": {{maxcombo}},
      "last_played": {{last_played}}
    },
    ...
  ],
  "modes": 
  [
    {
      "mode": "{{mode}}",
      "difficulties":
      [
        {
          "difficulty": "{{difficulty}}",
          "scores": 
          [
            {
              "score": {{score}},
              "clear_status": {{cleared}},
              "time_played": {{time_played}},
            },
            ...
          ]
        },
        ...
      ]
    },
    ...
  ]
}
```

### `/songs/scores/{id: song_id}/{mode: mode_name}/{difficulty: difficulty_name}` http.GET
List of users scores for a given song difficulty.

*headers*
```json
    "Authorization": "Bearer {{bearer_token}}"
```
*payload*
```json
{
  "order_by": "time_played" OPTIONAL
}
```
*response*
```json
{
  "song_id": "a1b2c3d4e5f6",
  "mode": "SINGLE",
  "difficulty": "EXPERT",
  "best_score": 
  {
    "best_score": 1000000,
    "lamp": "---",
    "rank": "AAA",
    "playcount": 3,
    "clearcount": 1,
    "maxcombo": 573,
    "last_played": 1234567890
  },
  "scores": 
  [
    {
      "score": 1000000,
      "clear_status": true,
      "time_played": 1234567890,
    },
    ...
  ]
}
```




## User endpoints: `/user`

### `/login` http.GET
Eagate login status for current authenticated user.

*headers*
```json
    "Authorization": "Bearer {{bearer_token}}"
```
*payload*
```json
{
}
```
*response*
```json
{
  "Name": "myusername@eagate.com",
  "NickName": "BstUser",
  "Cookie": "cookie",
  "Expiration": 1234567890,
  "WebUser":  "myusername@bst_web.com"
}

{
  "error": "an error message"
}
```

### `/login` http.POST
Link bst web user to eagate user.

*headers*
```json
    "Authorization": "Bearer {{bearer_token}}"
```
*payload*
```json
{
	"username": "myusername@eagate.com",
	"password": "MyPassword1!",
	"otp": "012345" OPTIONAL
}
```
*response*
```json
{
  "status": "ok"
}

{
  "error": "an error message"
}
```

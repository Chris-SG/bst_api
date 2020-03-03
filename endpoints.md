# Endpoints

## root endpoints: `/`

### GET `/status` ✅
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

### PATCH `/ddr/profile/update` ✅
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

### PATCH `/ddr/profile/refresh` ✅
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

### GET `/ddr/songs` ✅
List of songs currently in the database.

*headers*
```

```
*payload*
```json
{
  "order_by":
  [
    "Name",
    "Artist DESC",
    "Id ASC"
  ] OPTIONAL ❌
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

### PATCH `/ddr/songs` ✅
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

### GET `/ddr/songs/jackets` ✅
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

### GET `/ddr/songs/{id: song_id}` ✅
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

### GET `/ddr/songs/scores` ❌
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

### GET `/ddr/songs/scores/{id: song_id}` ❌
List of users scores for given song.

*headers*
```json
    "Authorization": "Bearer {{bearer_token}}"
```
*payload*
```json
{
  "order_by":
  [
    "mode",
    "difficulty DESC",
    "lamp ASC"
  ] OPTIONAL ❌
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

### GET `/ddr/songs/scores/{id: song_id}/{mode: mode_name}/{difficulty: difficulty_name}` ❌
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

### GET `/user/login` ✅
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

### POST `/user/login` ✅
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
	"otp": "012345" OPTIONAL ❌
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

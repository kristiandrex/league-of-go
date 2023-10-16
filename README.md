# League of Go

League of Go it's a simple Golang package for [League of React](https://github.com/kristiandrex/league-of-react) to download the champions information from [Data Dragon](https://developer.riotgames.com/docs/lol#data-dragon) and transform it to a custom structure.

This will return a JSON with this structure:

```json
{
  "version": "string",
  "champions": [
    "id": "string",
    "name": "string",
    "title": "string",
    "lore": "string",
    "thumbnail": "string",
    "skins": [
      "name": "string"
      "url": "string"
    ],
    "new": true
  ]
}
```

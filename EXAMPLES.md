# Examples

## Get latest tweets

```
$ pluck  -a 'data-aria-label-part="0"' -a '>' -d '</p>' -l -1 -u https://twitter.com/search\?f\=tweets\&vertical\=default\&q\=from%3ArealDonaldTrump\&src\=typd -s
```

## Get URL of original flickr image

```
$ pluck -a '"o":' -a 'display' -a ':"' -d '"' -u 'https://www.flickr.com/photos/mecedmonton/35612947565/in/album-72157683306176111/' -t
```

## Get URL of a 500px image

```
$ pluck -a 'twitter:title' -a "content='" -d "'" -u 'https://500px.com/photo/222234363/over-the-hills-by-bruno-pisani?ctx_page=1&from=editors' -s -t
```

## Read comments from Hacker News page

```
$ pluck -s -t -a 'class="c00"' -a '>' -d '</span' -u https://news.ycombinator.com/item\?id=14911467 | more
```


## Get Rotten Tomatoes movie scores

`rt.toml`:

```toml
[[pluck]]
activators = ["@context","schema.org",'"name":"']
deactivator = '"'
name = "name"

[[pluck]]
activators = ["all-critics-numbers","Reviews Counted:","<span>"]
deactivator = "<"
name = "critic_ratings"

[[pluck]]
activators = ["all-critics-numbers","Average Rating:",">"]
deactivator = "/"
name = "average_critic_rating"

[[pluck]]
activators = ["audience-score","Average Rating:",">"]
deactivator = "/"
name = "average_user_rating"

[[pluck]]
activators = ["User Ratings:",">"]
deactivator = "<"
name = "user_ratings"
```

```bash
$ pluck -s -c rt.toml -u https://www.rottentomatoes.com/m/spider_man_homecoming/
```

Returns:

```json
{
    "average_critic_rating": "7.6",
    "average_user_rating": "4.3",
    "critic_ratings": "276",
    "name": "Spider-Man: Homecoming",
    "user_ratings": "85,131"
}
```
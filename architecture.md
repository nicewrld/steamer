# steamer
steamer is my steam api scraper to build massive graphs of players. it is meant to be used in combination with some group of proxies, or in my case bright data because I think it's probably cheap enough to get away with, and saves me time.

I did this in 2021, but it was small and only ended up graphing like 50k of the accounts relationships. I want to be **bigger**, the number must grow.

## rough arch
```
  ┌─────────────────┐                                 
  │                 │                                 
  │  bright  data   │                                 
  │                 │                                 
  └─────────────────┘                                 
           ▲                                          
           │get proxy ips                             
           │                                          
  ┌────────┴────────┐           ┌──────────────────┐  
  │                 │   init    │                  │  
  │      core       ├──────────►│     database     │  
  │                 │           │                  │  
  └────────┬────────┘           └──────────────────┘  
           │                              ▲           
  create, one ip per                      │           
           ▼                              │           
  ┌─────────────────┐                     │           
  │                 │    get accounts to  │           
  │     workers     ├─────────────────────┘           
  │                 │    process, add                 
  └────────┬────────┘    updated accounts             
           │             and relationships            
        scrape                                        
           │                                          
           ▼                                          
  ┌─────────────────┐                                 
  │                 │                                 
  │     workers     │                                 
  │                 │                                 
  └─────────────────┘     
```

## pondering
looking at my steam account, i have 51 friends now, and my account will be the seed for this. if we get my account, we see a response like so:
```
{
  "friendslist": {
    "friends": [
      {
        "steamid": "76561197981746357",
        "relationship": "friend",
        "friend_since": 1434334976
      },
      // ... repeat until friends done
    ]
  }
}
```

when we call the steam api, our response for my account has 4211 bytes in the body, and 204 bytes in the header. I'm trying to get an understanding of what kind of costs this is gonna take, so we can take a response and filter it a bit to understand what the overhead cost of the 'scaffolding' of the api will be:

```
curl -s "http://api.steampowered.com/ISteamUser/GetFriendList/v1/?key=$STEAM_API_KEY&steamid=76561198024972007&relationship=friend" \
| jq -c 'del(.friendslist)' \
| wc -c

19
```

well that was a waste of time, its basically nothing, but that puts us at like 4192 bytes, and with 51 friends that's roughly 83 bytes per friend.

there will of course be bins that form based on number of friends, that we can use to roughly estimate the low and high ends:
```
10 friends: 1 kb
50 friends: 4 kb
100 friends: 8 kb
200 friends: 17 kb
300 friends: 25 kb
1000 friends: 83 kb
```
###### note: steam friends are limited to 250 (+5 per steam level after this)

i'm going to be billed by the GB, so i can get a good amount of data from a single GB, low enough that it doesn't matter to count. something like 12 million accounts or something, minus a bit for the headers. I think most people have between 25-100 friends, and at 200 bytes for headers thats like 5% overhead. so maybe closer to 11 million accounts per GB.

Steam limits us however to 100k requests per day per api key. I don't know if this is actually enforced per api key or per IP. I am going to assume this rate limit is enforced on an IP basis, and will be evading it. If it doesn't work, I will have to get more steam accounts.

# Real-time-forum

## Description
Real-time web forum that allows users to register, create posts, comment on posts, and send private chat messages to one another using Gorilla websocket

## Audit questions
[Link to audit questions](https://01.kood.tech/git/root/public/src/branch/master/subjects/real-time-forum/audit)

## Running

`go run .` from command line to run on port 8080  
For other port use `go run . --p=PORT_NR` or `go run . --port=PORT_NR`

## Dummy users for testing (With chat messages)

| Username | Email | Password |
| ----------- | ----------- | ----------- |
| test1 | test1@forum | test1 |
| test2 | test2@forum | test2 |
| traveler1994 |johnsmith@email.com | testtest |
| Hiker_Gal | jane@hike.com | testtest |
| GreenGuru | laura@flower.com | testtest |
| TechnoTina | tina@techno.com | testtest |

Between traveler1994 and Hiker_Gal there are 36 chat messages  
Between traveler1994 and GreenGuru there are 10 chat messages

## Authors
[obudarah / Olena Budarahina](https://01.kood.tech/git/obudarah)  
[ehspp / Erik Hans Sepp](https://01.kood.tech/git/ehspp)
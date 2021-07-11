# Goventy
This is a simple event management REST API built with Go (Mux & Gorm).
It allows you to register, create an event with different tickets & presenters.

## Routes
`POST` **/auth/register** - register on the platform

`POST` **/auth/login** - log into the platform

`GET` **/events** - get all events

`GET` **/events/{ID}** - get details about a particular event

`GET` **/events/{ID}/tickets** - get event tickets

GET` **/events/{ID}/presenters** - get event presenters

`POST` **/events** - create a new event (authenticated)

`PUT` **/events/{ID}** - update an event (authenticated)

`POST` **/tickets** - create a new ticket for an event (authenticated)

`PUT` **/tickets/{ID}** - update a ticket (authenticated)

`GET` **/tickets/{ID}** - get details about a particular ticket

`POST` **/presenters** - create a new presenter for an event (authenticated)

`PUT` **/presenters/{ID}** - update a presenter (authenticated)

`GET` **/presenters/{ID}** - get details about a particular presenter

`GET` **/tags** - get all tags
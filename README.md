title: postman
description:
authors: icy
categories: OP projects
created: 2022-12-06
updated: 2022-12-07
version: 0.0.18




##### EN | [RU](README_RU.md)


# Your friendly neighbour mailing list service

- Written in pure GO
- Celery client integration
- Email templates with gohtml syntax supported
- Check if email was read
- All around nice guy letting everybody know about latest gossip in town

## API methods

#### Create new mailing list    

`POST /api/list` - response is an ID for new mailing list

#### Add user to mailing list  

`POST /api/list/{id}` - where [id](id) is ID list's ID gotten from creation
  Request's body is a JSON struct with field names corresponding to user's information needed for Email template see [Add user](#add-user)

#### Add multiple users   

`POST /api/list/{id}/batch` - where [id](id) is ID list's ID gotten from creation
  Request's body is a JSON array with structs where any individual struct is the same as with single user addition

#### Get all the users in mailing list with specific ID   

`GET /api/list/{id}` - response is an array of JSON structs of individual user's information in mailing list with specific ID

#### Get read statistics from mailing list  

`GET api/list/{id}/stat` - get information about who read email from mailing list 
  Checkout implementation [Read statistics](#read-statistics)

#### Add email template to mailing list    

`POST /api/list/{id}/template` - request's body is a email template in text/html  
  [html/template](https://pkg.go.dev/html/template) is used for email template implementation. This means that all syntax features of gohtml format
  are available, see [Basic template for a mailing list](#basic-template-for-a-mailing-list)

## Examples

#### Add user

> `POST /api/list/{id}`
```
{
"email": "moya@govorit.net",
"name": "Alice"
"bonuses": "1200",
"link": "season.shop/id=4535345432"
}
```

#### Basic template for a mailing list

```html
<html>
<body>
     <h1>👋 Hello {{ .name}}</h1>
    <p>You have {{ .bonuses}} bonusSeas 🐳</p>
     <p>Spend them wisely on:
     <ul>
     <li>Whale rental 🐋</li>
     <li>Steamboat tickets 💨</li>
     <li>Investing in sea 🪣</li> 
     </ul>
     </p>
     <p>More inforamtion in your profile <a href="{{ .link}}"><span>here</span> </a></p>
</body>
</html>
## There are multiple better ways to construct an email HTML template, this is just a simple example
```

#### This example user information and email template produce the following email


> **To:** <em>moya@govorit.net</em>   
> **From:** _sea@son.shop_
<html>
  <body>
  <h1>👋 Hello Alice</h1>
  <p>You have 1200 bonusSeas 🐳</p>
  <p>Spend them wisely on:
  <ul>
  <li>Whale rental 🐋</li>
  <li>Steamboat tickets 💨</li>
  <li>Investing in sea 🪣</li> 
  </ul>
  </p>
  <p>More inforamtion in your profile <a href="season.shop"><span>here</span> </a></p>
  </body>
  </html>



## Read statistics

Checking if an email has been read is implemented using [RFC3798](https://datatracker.ietf.org/doc/html/rfc3798) header    
You can also use service handler `GET /api/{email}/read` to deliberately register that user with this `{email}` has read the message  

## Usage with celery

Function `postman.mail(id)` takes a mailing list [id](id) as an argument and starts constructing and sending emails based on template provided for mailing list with this ID
title: postman
description: 
authors: icy
categories: 
created: 2022-12-06
updated: 2022-12-07
version: 0.0.18



 Your friendly neighbour mailing list service
Written in pure GO
# API methods
- Create new mailing list  
  `POST /api/list` - response is an ID for new mailing list
- Add user to mailing list
  `POST /api/list/{id}` - where [id](id) is ID list's ID gotten from creation
  Request's body is a JSON struct with field names corresponding to user's information needed for Email template
- Add multiple users 
  `POST /api/list/{id}/batch` - where [id](id) is ID list's ID gotten from creation
  Request's body is a JSON array with structs where any individual struct is the same as with single user addition
- Get all the users in mailing list with specific ID 
  `GET /api/list/{id}` - response is an array of JSON structs of individual user's information in mailing list with specific ID
- Get read statistics from mailing list
  `GET api/list/{id}/stat` - get information about who read email from mailing list 
  Checkout implementation [here](## Read statistics)
- Add email template to mailing list
  `POST /api/list/{id}/template` - request's body is a email template in text/html  
  [html/template](https://pkg.go.dev/html/template) is used for email template implementation. This means that all syntax features of gohtml format
  are available
# Example
> Add user `POST` request
```
{
"email": "moya@govorit.net",
"name": "Alice"
"bonuses": "1200",
"link": "bonusshop.com/id=4535345432"
}
```
> Basic template for a mailing list
```javascript
<html>
<body>
Hello {{ .name}}

You have {{ .bonuses}} bonuses

You can spend 'em all here {{ .link}}
</body>
</html>
```
> This example user information and email template produce the following email
<body>
  Hello Alice

You have 1200 bonuses

You can spend 'em all here bonusshop.com/id=4535345432
</body>
# Read statistics 
Checking if an email has been read is implemented using [RFC3798](https://datatracker.ietf.org/doc/html/rfc3798) header  
You can also use service handler `GET /api/{email}/read` to deliberately register that user with this `{email}` has read the message

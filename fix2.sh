sed -i 's/func GetUser(/func GetUserFromRequest(/g' auth/session.go
sed -i 's/req.Session.Get("_auth_user_id")/req.Session.Get("_auth_user_id")\n\t\t\t\t/g' auth/session.go

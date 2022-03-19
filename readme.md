## Steps for setting up private repo

### Git
 - Set up SSH with the remote git
 - Set `GOPRIVATE` in `.pam_environment`
 ```
 GOPRIVATE=sabariram.com
 ```
 - Set the global config of git
 ```
 url.git@bitbucket.org:SabariramC/goserverbase.insteadof=https://sabariram.com/goserverbase
 ```
 - In the module that is going to use this package add the following in the go.mod file
 ```
 replace sabariram.com/goserverbase => sabariram.com/goserverbase.git <<tag>>
 ```
 eg:
 ```
 replace sabariram.com/goserverbase => sabariram.com/goserverbase.git v0.1.1
 ```
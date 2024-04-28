
## Steps for setting up private repo

### Git
 - Set up SSH with the remote git
 - Set `GOPRIVATE` in `.pam_environment`
 ```
 GOPRIVATE=sabariram.com
 ```
 - Set the global config of git
 ```
 url.git@bitbucket.org:SabariramC/goserverbase.insteadof=https://github.com/sabariramc/goserverbase/v5
 ```
 - In the module that is going to use this package add the following in the go.mod file
 ```
 replace github.com/sabariramc/goserverbase/v5 => github.com/sabariramc/goserverbase/v5.git <<tag>>
 ```
 eg:
 ```
 replace github.com/sabariramc/goserverbase/v5 => github.com/sabariramc/goserverbase/v5.git v0.1.1
 ```
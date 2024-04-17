export message="Log: Added file trace, update logging for panic recovery, support for unknown object logging"
export version="v5.0.21"
export branch="master"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push origin $branch
git push bitbucket $version
git push bitbucket $branch``
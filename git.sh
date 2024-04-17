export message="Log object type identifier fix for multi log object"
export version="v5.0.19"
export branch="master"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push origin $branch
git push bitbucket $version
git push bitbucket $branch``
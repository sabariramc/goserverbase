export message="Fixed bug in linking trace with kafka consumer"
export version="v5.0.7"
export branch="master"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push origin $branch
git push bitbucket $version
git push bitbucket $branch
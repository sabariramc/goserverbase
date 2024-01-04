export message="Fixed kafka consumer panic for non auto commit config"
export version="v4.7.2"
export branch="master"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push origin $branch
git push bitbucket $version
git push bitbucket $branch
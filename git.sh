export message="Reworked kafka health check intraction with shutdown and added WG for shutdown timer"
export version="v5.0.4"
export branch="master"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push origin $branch
git push bitbucket $version
git push bitbucket $branch
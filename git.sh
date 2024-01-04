export message="Fixed custom error print not parsing original error"
export version="v4.7.4"
export branch="master"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push origin $branch
git push bitbucket $version
git push bitbucket $branch
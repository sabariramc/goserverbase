export message="downgrading mongo driver"
export version="v4.14.2"
export branch="v4"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push origin $branch
git push bitbucket $version
git push bitbucket $branch
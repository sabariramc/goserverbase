export message="Streamlined config: config should not be passed as pointer"
export version="v5.1.4"
export branch="master"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push origin $branch
git push bitbucket $version
git push bitbucket $branch``
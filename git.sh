export message="Added utility method to SQS and SNS wrapper"
export version="v5.0.12"
export branch="master"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push origin $branch
git push bitbucket $version
git push bitbucket $branch
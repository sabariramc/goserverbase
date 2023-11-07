export message="Error handling update"
export version="v3.18.0.segmentio"
export branch="segmentio"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push origin $branch
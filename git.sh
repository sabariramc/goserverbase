export message="kafka buffer overflow"
export version="v3.18.8.segmentio"
export branch="segmentio"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push origin $branch
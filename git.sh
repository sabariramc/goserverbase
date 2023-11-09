export message="removed sasl config from kafka client"
export version="v3.18.5.segmentio"
export branch="segmentio"
git add .
git commit -m "$message"
git tag $version
git push origin $version
git push origin $branch
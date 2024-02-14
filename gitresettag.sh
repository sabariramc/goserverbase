export tagname="v5.0.2"
git push origin :$tagname
git push bitbucket :$tagname
git tag --delete $tagname
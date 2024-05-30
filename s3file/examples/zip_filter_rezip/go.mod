module zip_filter_rezip

go 1.18

replace github.com/tlelson/aws => /home/delubu/Code/aws

require (
	github.com/aws/aws-sdk-go v1.44.96
	github.com/tlelson/aws v0.0.0-00010101000000-000000000000
)

require github.com/jmespath/go-jmespath v0.4.0 // indirect

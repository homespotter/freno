build-prod:
	docker build \
		--build-arg AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID} \
		--build-arg AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY} \
		-t freno:latest .

push-prod: aws-login
	docker tag freno:latest 769352775470.dkr.ecr.us-east-1.amazonaws.com/freno:latest
	docker push 769352775470.dkr.ecr.us-east-1.amazonaws.com/freno:latest

aws-login:
	(aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 769352775470.dkr.ecr.us-east-1.amazonaws.com) || \
	(aws ecr get-login --no-include-email --region us-east-1 | bash)


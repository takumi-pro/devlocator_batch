REGION=asia-northeast1
IMAGE=$(REGION)-docker.pkg.dev/sigma-method-409207/devlocator-batch/devlocator_batch

image-build:
	docker image build -t $(IMAGE):latest --target production --platform linux/amd64 -f docker/Dockerfile .

image-push:
	docker image push $(IMAGE):latest

run:
	docker container run --name devlocator_batch $(IMAGE)

create-jobs:
	gcloud run jobs create devlocator-batch \
	--image=$(IMAGE) \
	--region=$(REGION)

run-jobs:
	gcloud run jobs execute devlocator_batch
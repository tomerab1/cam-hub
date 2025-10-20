.PHONY: model

model:
	python -m pip install openvino-dev
	omz_downloader --name person-detection-retail-0013 --output models --precisions FP16
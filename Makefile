delete:
	@kubectl delete -f deploy/crds/clock_v1_wallclock_crd.yaml
	@kubectl delete -f deploy/crds/clock_v1_timezones_crd.yaml
	@kubectl delete -f deploy/operator.yaml

create:
	@kubectl create -f deploy/crds/clock_v1_wallclock_crd.yaml
	@kubectl create -f deploy/crds/clock_v1_timezones_crd.yaml
	@kubectl create -f deploy/operator.yaml

redeploy:
	@operator-sdk build giorgosdi/wallclock-operator
	@docker push giorgosdi/wallclock-operator

tz:
	@kubectl apply -f deploy/crds/clock_v1_timezones_cr.yaml

restart: delete redeploy create tz
/*global angular*/
exports.uploader = ['$http', '$q', 'notification', '$state', function ($http, $q, notification, $state) {
    return {
        restrict: 'E',
        scope: {
            prefix: '&',
            suffix: '&'
        },
        link: function ($scope) {
            $scope.uploaded = false;

            $scope.onFileSelect = function ($files) {
                if (!$files || !$files.length) return;
                $scope.file = $files[0];
            };
            $scope.upload = function () {
                if (!$scope.file) return;

                // let pattern = /^#\/[a-zA-Z0-9\-_]+\/(edit|show)\/([0-9a-fA-F]+)$/;
                // let data = location.hash.match(pattern);
                // if (!data) return alert("Unable to detect the entry ID");

                // let id = data[2];
                // let id = generatePushID();
                let URL = $scope.prefix();// + id + $scope.suffix();

                let fd = new FormData();
                fd.append('file', $scope.file);
                return $http.post(URL, fd, {
                    transformRequest: angular.identity,
                    headers: { 'Content-Type': undefined }
                })
                    .then(function (res) {
                        $scope.file = null;
                        $state.reload();
                        notification.log(res.data.error || "Image uploaded", { addnCls: 'humane-flatty-success' });
                    })
                    .catch(function (res) {
                        notification.log(res.data.error || "Could not upload", { addnCls: 'humane-flatty-error' });
                    });
            };
        },
        template: `<div class="row">
				<style>
					.uploader {
						color: #333;
						background-color: #f7f7f7;
						display: inline-block;
						margin-bottom: 0;
						font-weight: 400;
						text-align: center;
						vertical-align: middle;
						touch-action: manipulation;
						background-image: none;
						cursor: pointer;
						border: 1px dashed #ccc;
						white-space: nowrap;
						padding: 24px 48px;
						font-size: 14px;
						line-height: 1.42857;
						border-radius: 4px;
						-webkit-user-select: none;
						-moz-user-select: none;
						-ms-user-select: none;
						user-select: none;
					}
					.uploader.bg-success {
						background-color: #dff0d8;
					}
					.uploader.bg-danger {
						background-color: #f2dede;
					}
				</style>
				<div class="col-md-4" ng-hide="file">
					<div class="uploader"
						ngf-drop
						ngf-select
						ngf-drag-over-class="{pattern: 'image/*', accept:'bg-success', reject:'bg-danger', delay:50}"
						ngf-pattern="image/*"
						ngf-max-total-size="'1MB'"
						ngf-change="onFileSelect($files)"
						ngf-multiple="false">Select an image or drop it here</div>
				</div>
				<div class="col-md-4" ng-show="file">
					<button type="button" class="btn btn-success btn-lg" ng-click="upload()">
						<span class="glyphicon glyphicon-upload"></span> Upload the image
					</button>
				</div>
		</div>`
    };
}];

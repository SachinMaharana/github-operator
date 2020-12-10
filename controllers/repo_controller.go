/*
Copyright 2020 SachinMaharana.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/go-github/v28/github"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	githubv1 "github.com/sachinmaharana/github-operator/api/v1"
	"github.com/sachinmaharana/github-operator/git"
)

// RepoReconciler reconciles a Repo object
type RepoReconciler struct {
	client.Client
	Log       logr.Logger
	Scheme    *runtime.Scheme
	GitClient git.Client
}

// +kubebuilder:rbac:groups=github.sachinmaharana.com,resources=repoes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=github.sachinmaharana.com,resources=repoes/status,verbs=get;update;patch

// Reconcile is ...
func (r *RepoReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var requeAfter time.Duration = 2 * time.Second
	ctx := context.Background()
	log := r.Log.WithValues("Repo", req.NamespacedName)

	var repo githubv1.Repo

	if err := r.Get(ctx, req.NamespacedName, &repo); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	organizationRepo := fmt.Sprintf("%s/%s", repo.Spec.Organization, repo.Name)
	log.Info("Found local repository,", "name", repo.Name)
	log.Info("check", "repo", organizationRepo)

	_, resp, err := r.GitClient.GetRepo(ctx, repo.Spec.Organization, repo.Name)
	if err != nil && isNotFound(resp) {
		log.Info("Repo Not Found", "creating", organizationRepo)
		if err := r.GitClient.CreateRepo(ctx, repo.Spec.Organization, &repo); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{RequeueAfter: requeAfter}, nil
	}

	if err != nil {
		return ctrl.Result{}, err
	}
	log.Info("found remote repository", "name", organizationRepo)

	return ctrl.Result{}, nil
}

// SetupWithManager ...
func (r *RepoReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&githubv1.Repo{}).
		Complete(r)
}

func isNotFound(r *github.Response) bool {
	if r != nil && r.StatusCode == http.StatusNotFound {
		return true
	}
	return false
}

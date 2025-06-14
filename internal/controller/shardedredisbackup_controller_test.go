/*
Copyright 2021.

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
	"testing"
	"time"

	"github.com/3scale-sre/basereconciler/reconciler"
	saasv1alpha1 "github.com/3scale-sre/saas-operator/api/v1alpha1"
	testutil "github.com/3scale-sre/saas-operator/test/util"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func TestShardedRedisBackupReconciler_reconcileBackupList(t *testing.T) {
	type args struct {
		instance *saasv1alpha1.ShardedRedisBackup
		nextRun  time.Time
		shards   []string
	}

	tests := []struct {
		name        string
		args        args
		wantChanged bool
		wantStatus  saasv1alpha1.ShardedRedisBackupStatus
		wantErr     bool
	}{
		{
			name: "List is empty, adds a backup",
			args: args{
				nextRun: testutil.MustParseRFC3339("2023-09-01T00:01:00Z"),
				instance: &saasv1alpha1.ShardedRedisBackup{
					Spec:   saasv1alpha1.ShardedRedisBackupSpec{HistoryLimit: ptr.To(int32(10))},
					Status: saasv1alpha1.ShardedRedisBackupStatus{},
				},
				shards: []string{"shard01", "shard02"},
			},
			wantChanged: true,
			wantStatus: saasv1alpha1.ShardedRedisBackupStatus{
				Backups: []saasv1alpha1.BackupStatus{
					{
						Shard:        "shard02",
						ScheduledFor: metav1.NewTime(testutil.MustParseRFC3339("2023-09-01T00:01:00Z")),
						Message:      "backup scheduled",
						State:        saasv1alpha1.BackupPendingState,
					},
					{
						Shard:        "shard01",
						ScheduledFor: metav1.NewTime(testutil.MustParseRFC3339("2023-09-01T00:01:00Z")),
						Message:      "backup scheduled",
						State:        saasv1alpha1.BackupPendingState,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "No changes",
			args: args{
				nextRun: testutil.MustParseRFC3339("2023-09-01T00:01:00Z"),
				instance: &saasv1alpha1.ShardedRedisBackup{
					Spec: saasv1alpha1.ShardedRedisBackupSpec{HistoryLimit: ptr.To(int32(10))},
					Status: saasv1alpha1.ShardedRedisBackupStatus{
						Backups: []saasv1alpha1.BackupStatus{
							{
								Shard:        "shard02",
								ScheduledFor: metav1.NewTime(testutil.MustParseRFC3339("2023-09-01T00:01:00Z")),
								Message:      "backup scheduled",
								State:        saasv1alpha1.BackupPendingState,
							},
							{
								Shard:        "shard01",
								ScheduledFor: metav1.NewTime(testutil.MustParseRFC3339("2023-09-01T00:01:00Z")),
								Message:      "backup scheduled",
								State:        saasv1alpha1.BackupPendingState,
							},
						}},
				},
				shards: []string{"shard01", "shard02"},
			},
			wantChanged: false,
			wantStatus: saasv1alpha1.ShardedRedisBackupStatus{
				Backups: []saasv1alpha1.BackupStatus{
					{
						Shard:        "shard02",
						ScheduledFor: metav1.NewTime(testutil.MustParseRFC3339("2023-09-01T00:01:00Z")),
						Message:      "backup scheduled",
						State:        saasv1alpha1.BackupPendingState,
					},
					{
						Shard:        "shard01",
						ScheduledFor: metav1.NewTime(testutil.MustParseRFC3339("2023-09-01T00:01:00Z")),
						Message:      "backup scheduled",
						State:        saasv1alpha1.BackupPendingState,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Adds new backups",
			args: args{
				nextRun: testutil.MustParseRFC3339("2023-09-01T00:02:00Z"),
				instance: &saasv1alpha1.ShardedRedisBackup{
					Spec: saasv1alpha1.ShardedRedisBackupSpec{HistoryLimit: ptr.To(int32(10))},
					Status: saasv1alpha1.ShardedRedisBackupStatus{
						Backups: []saasv1alpha1.BackupStatus{
							{
								Shard:        "shard02",
								ScheduledFor: metav1.NewTime(testutil.MustParseRFC3339("2023-09-01T00:01:00Z")),
								Message:      "backup scheduled",
								State:        saasv1alpha1.BackupPendingState,
							},
							{
								Shard:        "shard01",
								ScheduledFor: metav1.NewTime(testutil.MustParseRFC3339("2023-09-01T00:01:00Z")),
								Message:      "backup scheduled",
								State:        saasv1alpha1.BackupPendingState,
							},
						}},
				},
				shards: []string{"shard01", "shard02"},
			},
			wantChanged: true,
			wantStatus: saasv1alpha1.ShardedRedisBackupStatus{
				Backups: []saasv1alpha1.BackupStatus{
					{
						Shard:        "shard02",
						ScheduledFor: metav1.NewTime(testutil.MustParseRFC3339("2023-09-01T00:02:00Z")),
						Message:      "backup scheduled",
						State:        saasv1alpha1.BackupPendingState,
					},
					{
						Shard:        "shard01",
						ScheduledFor: metav1.NewTime(testutil.MustParseRFC3339("2023-09-01T00:02:00Z")),
						Message:      "backup scheduled",
						State:        saasv1alpha1.BackupPendingState,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ShardedRedisBackupReconciler{
				Reconciler: &reconciler.Reconciler{},
			}

			got := r.reconcileBackupList(context.TODO(), tt.args.instance, tt.args.nextRun, tt.args.shards)

			if diff := cmp.Diff(tt.args.instance.Status, tt.wantStatus); len(diff) > 0 {
				t.Errorf("ShardedRedisBackupReconciler.reconcileBackupList() = diff %v", diff)
			}

			if got != tt.wantChanged {
				t.Errorf("ShardedRedisBackupReconciler.reconcileBackupList() = %v, want %v", got, tt.wantChanged)
			}
		})
	}
}

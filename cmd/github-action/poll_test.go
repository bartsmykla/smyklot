package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/smykla-labs/smyklot/pkg/config"
	"github.com/smykla-labs/smyklot/pkg/github"
)

var _ = Describe("Poll Pending CI [Unit]", func() {
	Describe("parsePendingCILabel", func() {
		It("should parse smyklot:pending-ci label as merge method", func() {
			method, requiredOnly, label := parsePendingCILabel(github.LabelPendingCIMerge)
			Expect(method).To(Equal(github.MergeMethodMerge))
			Expect(requiredOnly).To(BeFalse())
			Expect(label).To(Equal(github.LabelPendingCIMerge))
		})

		It("should parse smyklot:pending-ci:squash label as squash method", func() {
			method, requiredOnly, label := parsePendingCILabel(github.LabelPendingCISquash)
			Expect(method).To(Equal(github.MergeMethodSquash))
			Expect(requiredOnly).To(BeFalse())
			Expect(label).To(Equal(github.LabelPendingCISquash))
		})

		It("should parse smyklot:pending-ci:rebase label as rebase method", func() {
			method, requiredOnly, label := parsePendingCILabel(github.LabelPendingCIRebase)
			Expect(method).To(Equal(github.MergeMethodRebase))
			Expect(requiredOnly).To(BeFalse())
			Expect(label).To(Equal(github.LabelPendingCIRebase))
		})

		It("should parse smyklot:pending-ci:required label as required merge method", func() {
			method, requiredOnly, label := parsePendingCILabel(github.LabelPendingCIMergeRequired)
			Expect(method).To(Equal(github.MergeMethodMerge))
			Expect(requiredOnly).To(BeTrue())
			Expect(label).To(Equal(github.LabelPendingCIMergeRequired))
		})

		It("should parse smyklot:pending-ci:squash:required label as required squash method", func() {
			method, requiredOnly, label := parsePendingCILabel(github.LabelPendingCISquashRequired)
			Expect(method).To(Equal(github.MergeMethodSquash))
			Expect(requiredOnly).To(BeTrue())
			Expect(label).To(Equal(github.LabelPendingCISquashRequired))
		})

		It("should parse smyklot:pending-ci:rebase:required label as required rebase method", func() {
			method, requiredOnly, label := parsePendingCILabel(github.LabelPendingCIRebaseRequired)
			Expect(method).To(Equal(github.MergeMethodRebase))
			Expect(requiredOnly).To(BeTrue())
			Expect(label).To(Equal(github.LabelPendingCIRebaseRequired))
		})

		It("should return empty string for non-pending-ci labels", func() {
			method, requiredOnly, label := parsePendingCILabel("some-other-label")
			Expect(method).To(Equal(github.MergeMethod("")))
			Expect(requiredOnly).To(BeFalse())
			Expect(label).To(BeEmpty())
		})

		It("should return empty string for reaction labels", func() {
			method, requiredOnly, label := parsePendingCILabel(github.LabelReactionApprove)
			Expect(method).To(Equal(github.MergeMethod("")))
			Expect(requiredOnly).To(BeFalse())
			Expect(label).To(BeEmpty())
		})
	})

	Describe("filterPendingCIPRs", func() {
		It("should filter PRs with pending-ci merge label", func() {
			prs := []map[string]interface{}{
				{
					"number": float64(1),
					"labels": []interface{}{
						map[string]interface{}{"name": github.LabelPendingCIMerge},
					},
				},
				{
					"number": float64(2),
					"labels": []interface{}{
						map[string]interface{}{"name": "other-label"},
					},
				},
			}

			result := filterPendingCIPRs(prs)
			Expect(result).To(HaveLen(1))
			Expect(extractPRNumber(result[0].prData)).To(Equal(1))
			Expect(result[0].method).To(Equal(github.MergeMethodMerge))
			Expect(result[0].label).To(Equal(github.LabelPendingCIMerge))
		})

		It("should filter PRs with pending-ci squash label", func() {
			prs := []map[string]interface{}{
				{
					"number": float64(1),
					"labels": []interface{}{
						map[string]interface{}{"name": github.LabelPendingCISquash},
					},
				},
			}

			result := filterPendingCIPRs(prs)
			Expect(result).To(HaveLen(1))
			Expect(result[0].method).To(Equal(github.MergeMethodSquash))
		})

		It("should filter PRs with pending-ci rebase label", func() {
			prs := []map[string]interface{}{
				{
					"number": float64(1),
					"labels": []interface{}{
						map[string]interface{}{"name": github.LabelPendingCIRebase},
					},
				},
			}

			result := filterPendingCIPRs(prs)
			Expect(result).To(HaveLen(1))
			Expect(result[0].method).To(Equal(github.MergeMethodRebase))
		})

		It("should return empty slice when no PRs have pending-ci labels", func() {
			prs := []map[string]interface{}{
				{
					"number": float64(1),
					"labels": []interface{}{
						map[string]interface{}{"name": "other-label"},
					},
				},
			}

			result := filterPendingCIPRs(prs)
			Expect(result).To(BeEmpty())
		})

		It("should handle PRs without labels field", func() {
			prs := []map[string]interface{}{
				{
					"number": float64(1),
				},
			}

			result := filterPendingCIPRs(prs)
			Expect(result).To(BeEmpty())
		})

		It("should handle PRs with empty labels array", func() {
			prs := []map[string]interface{}{
				{
					"number": float64(1),
					"labels": []interface{}{},
				},
			}

			result := filterPendingCIPRs(prs)
			Expect(result).To(BeEmpty())
		})

		It("should only pick first pending-ci label per PR", func() {
			prs := []map[string]interface{}{
				{
					"number": float64(1),
					"labels": []interface{}{
						map[string]interface{}{"name": github.LabelPendingCIMerge},
						map[string]interface{}{"name": github.LabelPendingCISquash},
					},
				},
			}

			result := filterPendingCIPRs(prs)
			Expect(result).To(HaveLen(1))
			// First label wins
			Expect(result[0].method).To(Equal(github.MergeMethodMerge))
		})

		It("should filter multiple PRs with different pending-ci labels", func() {
			prs := []map[string]interface{}{
				{
					"number": float64(1),
					"labels": []interface{}{
						map[string]interface{}{"name": github.LabelPendingCIMerge},
					},
				},
				{
					"number": float64(2),
					"labels": []interface{}{
						map[string]interface{}{"name": github.LabelPendingCISquash},
					},
				},
				{
					"number": float64(3),
					"labels": []interface{}{
						map[string]interface{}{"name": "other-label"},
					},
				},
			}

			result := filterPendingCIPRs(prs)
			Expect(result).To(HaveLen(2))
		})
	})

	Describe("extractPRNumber", func() {
		It("should extract PR number from PR data", func() {
			pr := map[string]interface{}{
				"number": float64(42),
			}
			Expect(extractPRNumber(pr)).To(Equal(42))
		})

		It("should return 0 for missing number field", func() {
			pr := map[string]interface{}{}
			Expect(extractPRNumber(pr)).To(Equal(0))
		})

		It("should return 0 for invalid number type", func() {
			pr := map[string]interface{}{
				"number": "not-a-number",
			}
			Expect(extractPRNumber(pr)).To(Equal(0))
		})
	})

	Describe("processPendingCIPR", func() {
		var server *httptest.Server

		AfterEach(func() {
			if server != nil {
				server.Close()
			}
		})

		Context("when CI is passing", func() {
			It("should merge the PR and remove label", func() {
				mergeRequested := false
				labelRemoved := false
				commentPosted := false

				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch {
					case r.URL.Path == "/repos/owner/repo/pulls/42" && r.Method == "GET":
						w.WriteHeader(http.StatusOK)
						_ = json.NewEncoder(w).Encode(map[string]interface{}{
							"head": map[string]interface{}{
								"sha": "abc123",
							},
						})

					case r.URL.Path == "/repos/owner/repo/commits/abc123/check-runs" && r.Method == "GET":
						w.WriteHeader(http.StatusOK)
						_ = json.NewEncoder(w).Encode(map[string]interface{}{
							"total_count": 2,
							"check_runs": []map[string]interface{}{
								{"status": "completed", "conclusion": "success"},
								{"status": "completed", "conclusion": "success"},
							},
						})

					case r.URL.Path == "/repos/owner/repo/pulls/42/merge" && r.Method == "PUT":
						mergeRequested = true
						w.WriteHeader(http.StatusOK)
						_ = json.NewEncoder(w).Encode(map[string]interface{}{
							"merged": true,
						})

					case r.URL.Path == "/repos/owner/repo/issues/42/labels/smyklot:pending-ci" && r.Method == "DELETE":
						labelRemoved = true
						w.WriteHeader(http.StatusOK)

					case r.URL.Path == "/repos/owner/repo/issues/42/comments" && r.Method == "POST":
						commentPosted = true
						w.WriteHeader(http.StatusCreated)
						_ = json.NewEncoder(w).Encode(map[string]interface{}{
							"id": 1,
						})

					default:
						w.WriteHeader(http.StatusNotFound)
					}
				}))

				client, err := github.NewClient("test-token", server.URL)
				Expect(err).NotTo(HaveOccurred())

				bc := config.Default()
				pr := pendingCIPR{
					prData: map[string]interface{}{"number": float64(42)},
					method: github.MergeMethodMerge,
					label:  github.LabelPendingCIMerge,
				}

				err = processPendingCIPR(context.Background(), client, bc, "owner", "repo", pr)
				Expect(err).NotTo(HaveOccurred())
				Expect(mergeRequested).To(BeTrue())
				Expect(labelRemoved).To(BeTrue())
				Expect(commentPosted).To(BeTrue())
			})
		})

		Context("when CI is failing", func() {
			It("should remove label and post failure comment", func() {
				labelRemoved := false
				commentPosted := false
				var postedComment string

				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch {
					case r.URL.Path == "/repos/owner/repo/pulls/42" && r.Method == "GET":
						w.WriteHeader(http.StatusOK)
						_ = json.NewEncoder(w).Encode(map[string]interface{}{
							"head": map[string]interface{}{
								"sha": "abc123",
							},
						})

					case r.URL.Path == "/repos/owner/repo/commits/abc123/check-runs" && r.Method == "GET":
						w.WriteHeader(http.StatusOK)
						_ = json.NewEncoder(w).Encode(map[string]interface{}{
							"total_count": 2,
							"check_runs": []map[string]interface{}{
								{"status": "completed", "conclusion": "success"},
								{"status": "completed", "conclusion": "failure"},
							},
						})

					case r.URL.Path == "/repos/owner/repo/issues/42/labels/smyklot:pending-ci" && r.Method == "DELETE":
						labelRemoved = true
						w.WriteHeader(http.StatusOK)

					case r.URL.Path == "/repos/owner/repo/issues/42/comments" && r.Method == "POST":
						commentPosted = true
						var body map[string]string
						_ = json.NewDecoder(r.Body).Decode(&body)
						postedComment = body["body"]
						w.WriteHeader(http.StatusCreated)
						_ = json.NewEncoder(w).Encode(map[string]interface{}{
							"id": 1,
						})

					default:
						w.WriteHeader(http.StatusNotFound)
					}
				}))

				client, err := github.NewClient("test-token", server.URL)
				Expect(err).NotTo(HaveOccurred())

				bc := config.Default()
				pr := pendingCIPR{
					prData: map[string]interface{}{"number": float64(42)},
					method: github.MergeMethodMerge,
					label:  github.LabelPendingCIMerge,
				}

				err = processPendingCIPR(context.Background(), client, bc, "owner", "repo", pr)
				Expect(err).NotTo(HaveOccurred())
				Expect(labelRemoved).To(BeTrue())
				Expect(commentPosted).To(BeTrue())
				Expect(postedComment).To(ContainSubstring("CI Failed"))
			})
		})

		Context("when CI is still pending", func() {
			It("should not merge or remove label", func() {
				mergeRequested := false
				labelRemoved := false

				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch {
					case r.URL.Path == "/repos/owner/repo/pulls/42" && r.Method == "GET":
						w.WriteHeader(http.StatusOK)
						_ = json.NewEncoder(w).Encode(map[string]interface{}{
							"head": map[string]interface{}{
								"sha": "abc123",
							},
						})

					case r.URL.Path == "/repos/owner/repo/commits/abc123/check-runs" && r.Method == "GET":
						w.WriteHeader(http.StatusOK)
						_ = json.NewEncoder(w).Encode(map[string]interface{}{
							"total_count": 2,
							"check_runs": []map[string]interface{}{
								{"status": "completed", "conclusion": "success"},
								{"status": "in_progress", "conclusion": nil},
							},
						})

					case r.URL.Path == "/repos/owner/repo/pulls/42/merge" && r.Method == "PUT":
						mergeRequested = true
						w.WriteHeader(http.StatusOK)

					case r.URL.Path == "/repos/owner/repo/issues/42/labels/smyklot:pending-ci" && r.Method == "DELETE":
						labelRemoved = true
						w.WriteHeader(http.StatusOK)

					default:
						w.WriteHeader(http.StatusNotFound)
					}
				}))

				client, err := github.NewClient("test-token", server.URL)
				Expect(err).NotTo(HaveOccurred())

				bc := config.Default()
				pr := pendingCIPR{
					prData: map[string]interface{}{"number": float64(42)},
					method: github.MergeMethodMerge,
					label:  github.LabelPendingCIMerge,
				}

				err = processPendingCIPR(context.Background(), client, bc, "owner", "repo", pr)
				Expect(err).NotTo(HaveOccurred())
				Expect(mergeRequested).To(BeFalse())
				Expect(labelRemoved).To(BeFalse())
			})
		})

		Context("when squash merge is requested", func() {
			It("should use squash merge method", func() {
				var mergeMethod string

				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch {
					case r.URL.Path == "/repos/owner/repo/pulls/42" && r.Method == "GET":
						w.WriteHeader(http.StatusOK)
						_ = json.NewEncoder(w).Encode(map[string]interface{}{
							"head": map[string]interface{}{
								"sha": "abc123",
							},
						})

					case r.URL.Path == "/repos/owner/repo/commits/abc123/check-runs" && r.Method == "GET":
						w.WriteHeader(http.StatusOK)
						_ = json.NewEncoder(w).Encode(map[string]interface{}{
							"total_count": 1,
							"check_runs": []map[string]interface{}{
								{"status": "completed", "conclusion": "success"},
							},
						})

					case r.URL.Path == "/repos/owner/repo/pulls/42/merge" && r.Method == "PUT":
						var body map[string]interface{}
						_ = json.NewDecoder(r.Body).Decode(&body)
						mergeMethod = body["merge_method"].(string)
						w.WriteHeader(http.StatusOK)
						_ = json.NewEncoder(w).Encode(map[string]interface{}{
							"merged": true,
						})

					case r.URL.Path == "/repos/owner/repo/issues/42/labels/smyklot:pending-ci:squash" && r.Method == "DELETE":
						w.WriteHeader(http.StatusOK)

					case r.URL.Path == "/repos/owner/repo/issues/42/comments" && r.Method == "POST":
						w.WriteHeader(http.StatusCreated)
						_ = json.NewEncoder(w).Encode(map[string]interface{}{
							"id": 1,
						})

					default:
						w.WriteHeader(http.StatusNotFound)
					}
				}))

				client, err := github.NewClient("test-token", server.URL)
				Expect(err).NotTo(HaveOccurred())

				bc := config.Default()
				pr := pendingCIPR{
					prData: map[string]interface{}{"number": float64(42)},
					method: github.MergeMethodSquash,
					label:  github.LabelPendingCISquash,
				}

				err = processPendingCIPR(context.Background(), client, bc, "owner", "repo", pr)
				Expect(err).NotTo(HaveOccurred())
				Expect(mergeMethod).To(Equal("squash"))
			})
		})
	})

	Describe("processPendingCIPRs", func() {
		var server *httptest.Server

		AfterEach(func() {
			if server != nil {
				server.Close()
			}
		})

		It("should return nil when no PRs have pending-ci labels", func() {
			prs := []map[string]interface{}{
				{
					"number": float64(1),
					"labels": []interface{}{
						map[string]interface{}{"name": "other-label"},
					},
				},
			}

			client, err := github.NewClient("test-token", "http://localhost")
			Expect(err).NotTo(HaveOccurred())

			bc := config.Default()
			err = processPendingCIPRs(context.Background(), client, bc, "owner", "repo", prs)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

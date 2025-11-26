package github_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bartsmykla/smyklot/pkg/github"
)

var _ = Describe("GitHub Client [Unit]", func() {
	var server *httptest.Server

	BeforeEach(func() {
		// Server will be set up in each test
	})

	AfterEach(func() {
		if server != nil {
			server.Close()
		}
	})

	Describe("NewClient", func() {
		It("should create a new client with token", func() {
			client, err := github.NewClient("test-token", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(client).NotTo(BeNil())
		})

		It("should create a new client with custom base URL", func() {
			client, err := github.NewClient("test-token", "https://api.github.com")
			Expect(err).NotTo(HaveOccurred())
			Expect(client).NotTo(BeNil())
		})

		It("should return error for empty token", func() {
			_, err := github.NewClient("", "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp(`(?i)empty.*token`))
		})
	})

	Describe("AddReaction", func() {
		Context("when adding reaction to a comment", func() {
			It("should add success reaction", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					Expect(r.Method).To(Equal("POST"))
					Expect(r.URL.Path).To(Equal("/repos/owner/repo/issues/comments/123/reactions"))
					Expect(r.Header.Get("Authorization")).To(Equal("token test-token"))

					var body map[string]string
					err := json.NewDecoder(r.Body).Decode(&body)
					Expect(err).NotTo(HaveOccurred())
					Expect(body["content"]).To(Equal("+1"))

					w.WriteHeader(http.StatusCreated)
					_ = json.NewEncoder(w).Encode(map[string]interface{}{
						"id":      1,
						"content": "+1",
					})
				}))

				client, err := github.NewClient("test-token", server.URL)
				Expect(err).NotTo(HaveOccurred())

				err = client.AddReaction(context.Background(), "owner", "repo", 123, github.ReactionSuccess)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should add error reaction", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					var body map[string]string
					_ = json.NewDecoder(r.Body).Decode(&body)
					Expect(body["content"]).To(Equal("-1"))

					w.WriteHeader(http.StatusCreated)
					_ = json.NewEncoder(w).Encode(map[string]interface{}{
						"id":      1,
						"content": "-1",
					})
				}))

				client, err := github.NewClient("test-token", server.URL)
				Expect(err).NotTo(HaveOccurred())

				err = client.AddReaction(context.Background(), "owner", "repo", 456, github.ReactionError)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should handle API errors", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusUnauthorized)
					_ = json.NewEncoder(w).Encode(map[string]string{
						"message": "Bad credentials",
					})
				}))

				client, err := github.NewClient("test-token", server.URL)
				Expect(err).NotTo(HaveOccurred())

				err = client.AddReaction(context.Background(), "owner", "repo", 123, github.ReactionSuccess)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("401"))
			})
		})
	})

	Describe("PostComment", func() {
		Context("when posting a comment on a PR", func() {
			It("should post a comment successfully", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					Expect(r.Method).To(Equal("POST"))
					Expect(r.URL.Path).To(Equal("/repos/owner/repo/issues/1/comments"))
					Expect(r.Header.Get("Authorization")).To(Equal("token test-token"))

					var body map[string]string
					err := json.NewDecoder(r.Body).Decode(&body)
					Expect(err).NotTo(HaveOccurred())
					Expect(body["body"]).To(Equal("Test comment"))

					w.WriteHeader(http.StatusCreated)
					_ = json.NewEncoder(w).Encode(map[string]interface{}{
						"id":   123,
						"body": "Test comment",
					})
				}))

				client, err := github.NewClient("test-token", server.URL)
				Expect(err).NotTo(HaveOccurred())

				err = client.PostComment(context.Background(), "owner", "repo", 1, "Test comment")
				Expect(err).NotTo(HaveOccurred())
			})

			It("should handle empty comment body", func() {
				client, err := github.NewClient("test-token", "")
				Expect(err).NotTo(HaveOccurred())

				err = client.PostComment(context.Background(), "owner", "repo", 1, "")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(MatchRegexp(`(?i)empty.*comment`))
			})

			It("should handle API errors", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusForbidden)
					_ = json.NewEncoder(w).Encode(map[string]string{
						"message": "Forbidden",
					})
				}))

				client, err := github.NewClient("test-token", server.URL)
				Expect(err).NotTo(HaveOccurred())

				err = client.PostComment(context.Background(), "owner", "repo", 1, "Test comment")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("ApprovePR", func() {
		Context("when approving a pull request", func() {
			It("should approve PR successfully", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					Expect(r.Method).To(Equal("POST"))
					Expect(r.URL.Path).To(Equal("/repos/owner/repo/pulls/1/reviews"))
					Expect(r.Header.Get("Authorization")).To(Equal("token test-token"))

					var body map[string]string
					err := json.NewDecoder(r.Body).Decode(&body)
					Expect(err).NotTo(HaveOccurred())
					Expect(body["event"]).To(Equal("APPROVE"))

					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(map[string]interface{}{
						"id":    1,
						"state": "APPROVED",
					})
				}))

				client, err := github.NewClient("test-token", server.URL)
				Expect(err).NotTo(HaveOccurred())

				err = client.ApprovePR(context.Background(), "owner", "repo", 1)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should handle API errors", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusUnprocessableEntity)
					_ = json.NewEncoder(w).Encode(map[string]string{
						"message": "Pull request already approved",
					})
				}))

				client, err := github.NewClient("test-token", server.URL)
				Expect(err).NotTo(HaveOccurred())

				err = client.ApprovePR(context.Background(), "owner", "repo", 1)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("MergePR", func() {
		Context("when merging a pull request", func() {
			It("should merge PR successfully", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					Expect(r.Method).To(Equal("PUT"))
					Expect(r.URL.Path).To(Equal("/repos/owner/repo/pulls/1/merge"))
					Expect(r.Header.Get("Authorization")).To(Equal("token test-token"))

					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(map[string]interface{}{
						"sha":     "abc123",
						"merged":  true,
						"message": "Pull Request successfully merged",
					})
				}))

				client, err := github.NewClient("test-token", server.URL)
				Expect(err).NotTo(HaveOccurred())

				err = client.MergePR(context.Background(), "owner", "repo", 1, github.MergeMethodMerge)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should handle merge conflicts", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusConflict)
					_ = json.NewEncoder(w).Encode(map[string]string{
						"message": "Merge conflict",
					})
				}))

				client, err := github.NewClient("test-token", server.URL)
				Expect(err).NotTo(HaveOccurred())

				err = client.MergePR(context.Background(), "owner", "repo", 1, github.MergeMethodMerge)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("409"))
			})

			It("should handle PR not mergeable", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusMethodNotAllowed)
					_ = json.NewEncoder(w).Encode(map[string]string{
						"message": "Pull Request is not mergeable",
					})
				}))

				client, err := github.NewClient("test-token", server.URL)
				Expect(err).NotTo(HaveOccurred())

				err = client.MergePR(context.Background(), "owner", "repo", 1, github.MergeMethodMerge)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("GetPRInfo", func() {
		Context("when getting PR information", func() {
			It("should get PR info successfully", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					defer GinkgoRecover()

					Expect(r.Method).To(Equal("GET"))
					Expect(r.Header.Get("Authorization")).To(Equal("token test-token"))

					// Handle both PR info and reviews requests
					switch r.URL.Path {
					case "/repos/owner/repo/pulls/1":
						w.WriteHeader(http.StatusOK)
						_ = json.NewEncoder(w).Encode(map[string]interface{}{
							"number":    1,
							"state":     "open",
							"mergeable": true,
							"title":     "Test PR",
							"body":      "Test description",
							"user": map[string]interface{}{
								"login": "testuser",
							},
						})
					case "/repos/owner/repo/pulls/1/reviews":
						w.WriteHeader(http.StatusOK)
						_ = json.NewEncoder(w).Encode([]map[string]interface{}{
							{
								"state": "APPROVED",
								"user": map[string]interface{}{
									"login": "reviewer1",
								},
							},
						})
					default:
						Fail("unexpected request path: " + r.URL.Path)
					}
				}))

				client, err := github.NewClient("test-token", server.URL)
				Expect(err).NotTo(HaveOccurred())

				info, err := client.GetPRInfo(context.Background(), "owner", "repo", 1)
				Expect(err).NotTo(HaveOccurred())
				Expect(info).NotTo(BeNil())
				Expect(info.Number).To(Equal(1))
				Expect(info.State).To(Equal("open"))
				Expect(info.Mergeable).To(BeTrue())
				Expect(info.Title).To(Equal("Test PR"))
				Expect(info.Author).To(Equal("testuser"))
				Expect(info.ApprovedBy).To(ConsistOf("reviewer1"))
			})

			It("should handle PR not found", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
					_ = json.NewEncoder(w).Encode(map[string]string{
						"message": "Not Found",
					})
				}))

				client, err := github.NewClient("test-token", server.URL)
				Expect(err).NotTo(HaveOccurred())

				_, err = client.GetPRInfo(context.Background(), "owner", "repo", 999)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("404"))
			})
		})
	})

	Describe("IsTeamMember", func() {
		Context("when checking team membership", func() {
			It("should return true when user is a team member", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					Expect(r.Method).To(Equal("GET"))
					Expect(r.URL.Path).To(Equal("/orgs/test-org/teams/test-team/memberships/testuser"))
					Expect(r.Header.Get("Authorization")).To(Equal("token test-token"))

					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(map[string]interface{}{
						"state": "active",
						"role":  "member",
					})
				}))

				client, err := github.NewClient("test-token", server.URL)
				Expect(err).NotTo(HaveOccurred())

				isMember, err := client.IsTeamMember(context.Background(), "test-org", "test-team", "testuser")
				Expect(err).NotTo(HaveOccurred())
				Expect(isMember).To(BeTrue())
			})

			It("should return false when user is not a team member", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
					_ = json.NewEncoder(w).Encode(map[string]string{
						"message": "Not Found",
					})
				}))

				client, err := github.NewClient("test-token", server.URL)
				Expect(err).NotTo(HaveOccurred())

				isMember, err := client.IsTeamMember(context.Background(), "test-org", "test-team", "nonmember")
				Expect(err).NotTo(HaveOccurred())
				Expect(isMember).To(BeFalse())
			})

			It("should return false when team membership is pending", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(map[string]interface{}{
						"state": "pending",
						"role":  "member",
					})
				}))

				client, err := github.NewClient("test-token", server.URL)
				Expect(err).NotTo(HaveOccurred())

				isMember, err := client.IsTeamMember(context.Background(), "test-org", "test-team", "pendinguser")
				Expect(err).NotTo(HaveOccurred())
				Expect(isMember).To(BeFalse())
			})

			It("should handle API errors", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusForbidden)
					_ = json.NewEncoder(w).Encode(map[string]string{
						"message": "Forbidden",
					})
				}))

				client, err := github.NewClient("test-token", server.URL)
				Expect(err).NotTo(HaveOccurred())

				_, err = client.IsTeamMember(context.Background(), "test-org", "test-team", "testuser")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Error Handling", func() {
		Context("when handling various error conditions", func() {
			It("should handle network errors", func() {
				// Create client with invalid URL
				client, err := github.NewClient("test-token", "http://invalid-url-that-does-not-exist.local")
				Expect(err).NotTo(HaveOccurred())

				err = client.AddReaction(context.Background(), "owner", "repo", 1, github.ReactionSuccess)
				Expect(err).To(HaveOccurred())
			})

			It("should handle malformed JSON responses", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte("invalid json"))
				}))

				client, err := github.NewClient("test-token", server.URL)
				Expect(err).NotTo(HaveOccurred())

				_, err = client.GetPRInfo(context.Background(), "owner", "repo", 1)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})

package clients

import pb "github.com/LimeOnTop/interverse-contracts/user/gen"

func NewFeedbackClient(userServiceURL string) pb.FeedbackServiceClient {
	return pb.NewFeedbackServiceClient(dialGRPC(userServiceURL, "user-service (feedback)"))
}

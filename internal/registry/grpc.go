package registry

import (
	"context"

	pb "github.com/Sithumli/Beacon/api/proto"
	"github.com/Sithumli/Beacon/internal/core"
	"github.com/Sithumli/Beacon/internal/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GRPCServer implements the RegistryService gRPC server
type GRPCServer struct {
	pb.UnimplementedRegistryServiceServer
	service *Service
}

// NewGRPCServer creates a new gRPC server for the registry
func NewGRPCServer(svc *Service) *GRPCServer {
	return &GRPCServer{service: svc}
}

// Register registers a new agent
func (s *GRPCServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	agent := &core.Agent{
		Name:        req.Name,
		Version:     req.Version,
		Description: req.Description,
		Endpoint: core.Endpoint{
			Host:     req.Endpoint.GetHost(),
			Port:     int(req.Endpoint.GetPort()),
			Protocol: req.Endpoint.GetProtocol(),
		},
		Capabilities: make([]core.Capability, len(req.Capabilities)),
		Metadata: core.Metadata{
			Author: req.Metadata.GetAuthor(),
			Tags:   req.Metadata.GetTags(),
		},
	}

	for i, cap := range req.Capabilities {
		agent.Capabilities[i] = core.Capability{
			Name:        cap.Name,
			Description: cap.Description,
		}
	}

	registered, err := s.service.Register(ctx, agent)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "registration failed: %v", err)
	}

	return &pb.RegisterResponse{
		AgentId: registered.ID,
		Agent:   agentToProto(registered),
	}, nil
}

// Deregister removes an agent
func (s *GRPCServer) Deregister(ctx context.Context, req *pb.DeregisterRequest) (*pb.DeregisterResponse, error) {
	err := s.service.Deregister(ctx, req.AgentId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "agent not found: %v", err)
	}
	return &pb.DeregisterResponse{Success: true}, nil
}

// GetAgent retrieves an agent by ID
func (s *GRPCServer) GetAgent(ctx context.Context, req *pb.GetAgentRequest) (*pb.GetAgentResponse, error) {
	agent, err := s.service.GetAgent(ctx, req.AgentId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "agent not found: %v", err)
	}
	return &pb.GetAgentResponse{Agent: agentToProto(agent)}, nil
}

// ListAgents lists all agents
func (s *GRPCServer) ListAgents(ctx context.Context, req *pb.ListAgentsRequest) (*pb.ListAgentsResponse, error) {
	// Build filter from request
	filter := &store.AgentFilter{}
	if req.Status != pb.AgentStatus_AGENT_STATUS_UNSPECIFIED {
		status := protoToAgentStatus(req.Status)
		filter.Status = &status
	}
	if req.Capability != "" {
		filter.Capability = &req.Capability
	}
	if len(req.Tags) > 0 {
		filter.Tags = req.Tags
	}

	agents, err := s.service.ListAgents(ctx, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list agents: %v", err)
	}

	pbAgents := make([]*pb.Agent, len(agents))
	for i, agent := range agents {
		pbAgents[i] = agentToProto(agent)
	}

	return &pb.ListAgentsResponse{Agents: pbAgents}, nil
}

// Discover finds agents by capability
func (s *GRPCServer) Discover(ctx context.Context, req *pb.DiscoverRequest) (*pb.DiscoverResponse, error) {
	agents, err := s.service.Discover(ctx, req.Capability)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "discovery failed: %v", err)
	}

	pbAgents := make([]*pb.Agent, len(agents))
	for i, agent := range agents {
		pbAgents[i] = agentToProto(agent)
	}

	return &pb.DiscoverResponse{Agents: pbAgents}, nil
}

// Heartbeat updates agent health
func (s *GRPCServer) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	err := s.service.Heartbeat(ctx, req.AgentId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "agent not found: %v", err)
	}
	return &pb.HeartbeatResponse{
		Success: true,
		Status:  pb.AgentStatus_AGENT_STATUS_HEALTHY,
	}, nil
}

// Watch streams agent changes
func (s *GRPCServer) Watch(req *pb.WatchRequest, stream pb.RegistryService_WatchServer) error {
	ctx := stream.Context()

	// Subscribe to registry events
	watchID, eventCh := s.service.Watch(req.Capabilities)
	defer s.service.Unwatch(watchID)

	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-eventCh:
			if !ok {
				return nil
			}
			resp := &pb.WatchResponse{
				Event: watchEventTypeToProto(event.Type),
				Agent: agentToProto(event.Agent),
			}
			if err := stream.Send(resp); err != nil {
				return err
			}
		}
	}
}

func agentToProto(a *core.Agent) *pb.Agent {
	caps := make([]*pb.Capability, len(a.Capabilities))
	for i, cap := range a.Capabilities {
		caps[i] = &pb.Capability{
			Name:        cap.Name,
			Description: cap.Description,
		}
	}

	return &pb.Agent{
		Id:          a.ID,
		Name:        a.Name,
		Version:     a.Version,
		Description: a.Description,
		Endpoint: &pb.Endpoint{
			Host:     a.Endpoint.Host,
			Port:     int32(a.Endpoint.Port),
			Protocol: a.Endpoint.Protocol,
		},
		Capabilities: caps,
		Metadata: &pb.Metadata{
			Author: a.Metadata.Author,
			Tags:   a.Metadata.Tags,
		},
		Status:        statusToProto(a.Status),
		RegisteredAt:  timestamppb.New(a.RegisteredAt),
		LastHeartbeat: timestamppb.New(a.LastHeartbeat),
	}
}

func statusToProto(s core.AgentStatus) pb.AgentStatus {
	switch s {
	case core.StatusHealthy:
		return pb.AgentStatus_AGENT_STATUS_HEALTHY
	case core.StatusUnhealthy:
		return pb.AgentStatus_AGENT_STATUS_UNHEALTHY
	default:
		return pb.AgentStatus_AGENT_STATUS_UNKNOWN
	}
}

func protoToAgentStatus(s pb.AgentStatus) core.AgentStatus {
	switch s {
	case pb.AgentStatus_AGENT_STATUS_HEALTHY:
		return core.StatusHealthy
	case pb.AgentStatus_AGENT_STATUS_UNHEALTHY:
		return core.StatusUnhealthy
	default:
		return core.StatusUnknown
	}
}

func watchEventTypeToProto(t EventType) pb.WatchResponse_EventType {
	switch t {
	case EventRegistered:
		return pb.WatchResponse_EVENT_TYPE_REGISTERED
	case EventDeregistered:
		return pb.WatchResponse_EVENT_TYPE_DEREGISTERED
	case EventUpdated:
		return pb.WatchResponse_EVENT_TYPE_UPDATED
	case EventHealthChanged:
		return pb.WatchResponse_EVENT_TYPE_HEALTH_CHANGED
	default:
		return pb.WatchResponse_EVENT_TYPE_UNSPECIFIED
	}
}

package broker

import (
	"context"
	"encoding/json"

	"github.com/Sithumli/Beacon/internal/core"
	"github.com/Sithumli/Beacon/internal/store"
	pb "github.com/Sithumli/Beacon/api/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GRPCServer implements the BrokerService gRPC server
type GRPCServer struct {
	pb.UnimplementedBrokerServiceServer
	service *Service
}

// NewGRPCServer creates a new gRPC server for the broker
func NewGRPCServer(svc *Service) *GRPCServer {
	return &GRPCServer{service: svc}
}

// SendTask sends a task to a specific agent
func (s *GRPCServer) SendTask(ctx context.Context, req *pb.SendTaskRequest) (*pb.SendTaskResponse, error) {
	task, err := s.service.SendTask(ctx, req.FromAgent, req.ToAgent, req.Capability, req.Payload)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to send task: %v", err)
	}

	return &pb.SendTaskResponse{
		TaskId: task.ID,
		Task:   taskToProto(task),
	}, nil
}

// GetTask retrieves a task by ID
func (s *GRPCServer) GetTask(ctx context.Context, req *pb.GetTaskRequest) (*pb.GetTaskResponse, error) {
	task, err := s.service.GetTask(ctx, req.TaskId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "task not found: %v", err)
	}

	return &pb.GetTaskResponse{Task: taskToProto(task)}, nil
}

// UpdateTask updates a task's status/result
func (s *GRPCServer) UpdateTask(ctx context.Context, req *pb.UpdateTaskRequest) (*pb.UpdateTaskResponse, error) {
	taskStatus := protoToTaskStatus(req.Status)
	task, err := s.service.UpdateTask(ctx, req.TaskId, taskStatus, req.Result, req.Error)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to update task: %v", err)
	}

	return &pb.UpdateTaskResponse{Task: taskToProto(task)}, nil
}

// ListTasks lists tasks with filters
func (s *GRPCServer) ListTasks(ctx context.Context, req *pb.ListTasksRequest) (*pb.ListTasksResponse, error) {
	filter := &store.TaskFilter{}

	if req.Status != pb.TaskStatus_TASK_STATUS_UNSPECIFIED {
		status := protoToTaskStatus(req.Status)
		filter.Status = &status
	}
	if req.FromAgent != "" {
		filter.FromAgent = &req.FromAgent
	}
	if req.ToAgent != "" {
		filter.ToAgent = &req.ToAgent
	}
	if req.Limit > 0 {
		limit := int(req.Limit)
		filter.Limit = &limit
	}
	if req.Offset > 0 {
		offset := int(req.Offset)
		filter.Offset = &offset
	}

	tasks, err := s.service.ListTasks(ctx, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list tasks: %v", err)
	}

	pbTasks := make([]*pb.Task, len(tasks))
	for i, task := range tasks {
		pbTasks[i] = taskToProto(task)
	}

	return &pb.ListTasksResponse{
		Tasks: pbTasks,
		Total: int32(len(tasks)),
	}, nil
}

// CancelTask cancels a pending or running task
func (s *GRPCServer) CancelTask(ctx context.Context, req *pb.CancelTaskRequest) (*pb.CancelTaskResponse, error) {
	task, err := s.service.CancelTask(ctx, req.TaskId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to cancel task: %v", err)
	}

	return &pb.CancelTaskResponse{
		Success: true,
		Task:    taskToProto(task),
	}, nil
}

// Subscribe streams tasks for an agent
func (s *GRPCServer) Subscribe(req *pb.SubscribeRequest, stream pb.BrokerService_SubscribeServer) error {
	ctx := stream.Context()

	// Subscribe to task events
	subID, eventCh := s.service.Subscribe(req.AgentId)
	defer s.service.Unsubscribe(subID)

	// First, send any pending tasks
	pending, err := s.service.GetPendingTasks(ctx, req.AgentId)
	if err == nil {
		for _, task := range pending {
			resp := &pb.TaskEvent{
				Event: pb.TaskEvent_EVENT_TYPE_NEW_TASK,
				Task:  taskToProto(task),
			}
			if err := stream.Send(resp); err != nil {
				return err
			}
		}
	}

	// Then stream new events
	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-eventCh:
			if !ok {
				return nil
			}
			resp := &pb.TaskEvent{
				Event: taskEventTypeToProto(event.Type),
				Task:  taskToProto(event.Task),
			}
			if err := stream.Send(resp); err != nil {
				return err
			}
		}
	}
}

// RouteTask routes a task to any agent with the capability
func (s *GRPCServer) RouteTask(ctx context.Context, req *pb.RouteTaskRequest) (*pb.RouteTaskResponse, error) {
	task, err := s.service.RouteTask(ctx, req.FromAgent, req.Capability, req.Payload)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to route task: %v", err)
	}

	return &pb.RouteTaskResponse{
		TaskId:  task.ID,
		ToAgent: task.ToAgent,
		Task:    taskToProto(task),
	}, nil
}

func taskToProto(t *core.Task) *pb.Task {
	pbTask := &pb.Task{
		Id:         t.ID,
		FromAgent:  t.FromAgent,
		ToAgent:    t.ToAgent,
		Capability: t.Capability,
		Payload:    t.Payload,
		Status:     taskStatusToProto(t.Status),
		Result:     t.Result,
		Error:      t.Error,
		CreatedAt:  timestamppb.New(t.CreatedAt),
		UpdatedAt:  timestamppb.New(t.UpdatedAt),
	}

	if t.CompletedAt != nil {
		pbTask.CompletedAt = timestamppb.New(*t.CompletedAt)
	}

	return pbTask
}

func taskStatusToProto(s core.TaskStatus) pb.TaskStatus {
	switch s {
	case core.TaskPending:
		return pb.TaskStatus_TASK_STATUS_PENDING
	case core.TaskRunning:
		return pb.TaskStatus_TASK_STATUS_RUNNING
	case core.TaskCompleted:
		return pb.TaskStatus_TASK_STATUS_COMPLETED
	case core.TaskFailed:
		return pb.TaskStatus_TASK_STATUS_FAILED
	case core.TaskCancelled:
		return pb.TaskStatus_TASK_STATUS_CANCELLED
	default:
		return pb.TaskStatus_TASK_STATUS_UNSPECIFIED
	}
}

func protoToTaskStatus(s pb.TaskStatus) core.TaskStatus {
	switch s {
	case pb.TaskStatus_TASK_STATUS_PENDING:
		return core.TaskPending
	case pb.TaskStatus_TASK_STATUS_RUNNING:
		return core.TaskRunning
	case pb.TaskStatus_TASK_STATUS_COMPLETED:
		return core.TaskCompleted
	case pb.TaskStatus_TASK_STATUS_FAILED:
		return core.TaskFailed
	case pb.TaskStatus_TASK_STATUS_CANCELLED:
		return core.TaskCancelled
	default:
		return core.TaskPending
	}
}

func taskEventTypeToProto(t EventType) pb.TaskEvent_EventType {
	switch t {
	case EventNewTask:
		return pb.TaskEvent_EVENT_TYPE_NEW_TASK
	case EventTaskUpdated:
		return pb.TaskEvent_EVENT_TYPE_TASK_UPDATED
	case EventTaskCancelled:
		return pb.TaskEvent_EVENT_TYPE_TASK_CANCELLED
	default:
		return pb.TaskEvent_EVENT_TYPE_UNSPECIFIED
	}
}

// Helper for agents to execute tasks - converts JSON payload for handler
func UnmarshalPayload(payload []byte, v interface{}) error {
	return json.Unmarshal(payload, v)
}

func MarshalResult(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

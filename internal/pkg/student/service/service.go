package service

import (
	"context"
	stuPb "gocache/api/studentpb"
	"gocache/internal/pkg/student/dao"
	"gocache/internal/pkg/student/encode"
)

type StudentSrv struct {
	stuPb.UnimplementedStudentServiceServer
}

func (s *StudentSrv) StudentLogin(ctx context.Context, req *stuPb.StudentRequest) (resp *stuPb.StudentDetailResponse, err error) {
	resp = new(stuPb.StudentDetailResponse)
	resp.Code = encode.SUCCESS
	r, err := dao.NewStudentDao(ctx).ShowStudentInfo(req)
	if err != nil {
		resp.Code = encode.ERROR
		return
	}
	resp.StudentDetail = &stuPb.StudentResponse{
		Name:  r.Name,
		Score: float32(r.Score),
	}
	return
}

func (s *StudentSrv) StudentRegister(ctx context.Context, req *stuPb.StudentRequest) (resp *stuPb.StudentCommonResponse, err error) {
	resp = new(stuPb.StudentCommonResponse)
	resp.Code = encode.SUCCESS
	err = dao.NewStudentDao(ctx).CreateStudent(req)
	if err != nil {
		resp.Code = encode.ERROR
		resp.Message = encode.GetMsg(int(resp.Code))
		return
	}
	resp.Message = encode.GetMsg(int(resp.Code))
	return
}

func (s *StudentSrv) StudentLogout(ctx context.Context, req *stuPb.StudentRequest) (resp *stuPb.StudentCommonResponse, err error) {
	resp = new(stuPb.StudentCommonResponse)
	resp.Code = encode.SUCCESS
	resp.Message = encode.GetMsg(int(resp.Code))
	return
}

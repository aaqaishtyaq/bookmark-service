package v1

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	v1 "github.com/aaqaishtyaq/bookmark-service/pkg/api/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const apiVersion string = "v1"

// BookmarkServiceServer
type BookmarkServiceServer struct {
	db *sql.DB
	v1.UnimplementedBookmarkServiceServer
}

type BookmarkService interface {
	CreateBookmark(context.Context, *v1.CreateBookmarkReq) (*v1.CreateBookmarkRes, error)
	ReadBookmark(ctx context.Context, req *v1.ReadBookmarkReq) (*v1.ReadBookmarkRes, error)
	ListBookmarks(context.Context, *v1.ListBookmarksReq) (*v1.ListBookmarksRes, error)
	DeleteBookmark(ctx context.Context, req *v1.DeleteBookmarkReq) (*v1.DeleteBookmarkRes, error)
	mustEmbedUnimplementedBookmarkServiceServer()
}

// NewBookmarkServiceServer creates a Bookmark service
func NewBookmarkServiceServer(db *sql.DB) v1.BookmarkServiceServer {
	return &BookmarkServiceServer{db: db}
}

// checkAPI checks if the API version requested by client is supported by server
func (s *BookmarkServiceServer) checkAPI(api string) error {
	// API version is "" means use current version of the service
	if len(api) > 0 {
		if apiVersion != api {
			return status.Errorf(codes.Unimplemented,
				"unsupported API version: service implements API version '%s', but asked for '%s'", apiVersion, api)
		}
	}
	return nil
}

// connect returns SQL database connection from the pool
func (s *BookmarkServiceServer) connect(ctx context.Context) (*sql.Conn, error) {
	c, err := s.db.Conn(ctx)
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to connect to database "+err.Error())
	}
	return c, nil
}

// Create new Bookmar task
func (s *BookmarkServiceServer) CreateBookmark(ctx context.Context, req *v1.CreateBookmarkReq) (*v1.CreateBookmarkRes, error) {
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}

	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	// url, err := ptypes.Timestamp()
	// if err != nil {
	// 	return nil, status.Error(codes.InvalidArgument, "url field has invalid format-> "+err.Error())
	// }

	// insert Todo entity data
	var id int64
	c.QueryRowContext(ctx, `INSERT INTO bookmarks (url) VALUES ($1) RETURNING id`,
		req.Bookmark.Url).Scan(&id)

	if id == 0 {
		return nil, status.Error(codes.Unknown, "failed to insert into todo")
	}

	bk := &v1.Bookmark{
		Id:  id,
		Url: req.Bookmark.Url,
	}

	return &v1.CreateBookmarkRes{
		Api:      apiVersion,
		Bookmark: bk,
	}, nil
}

// ListBook
func (s *BookmarkServiceServer) ListBookmarks(ctx context.Context, req *v1.ListBookmarksReq) (*v1.ListBookmarksRes, error) {
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}

	// get SQL connection from pool
	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	// get Todo list
	rows, err := c.QueryContext(ctx, "select id, url from bookmarks")
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to select from Bookmarks "+err.Error())
	}

	defer rows.Close()

	bookmarks := []*v1.Bookmark{}
	for rows.Next() {
		bk := new(v1.Bookmark)
		if err := rows.Scan(&bk.Id, &bk.Url); err != nil {
			return nil, status.Error(codes.Unknown, "failed to retrieve field values from Bookmark row-> "+err.Error())
		}

		bookmarks = append(bookmarks, bk)
	}

	if err := rows.Err(); err != nil {
		return nil, status.Error(codes.Unknown, "failed to retrieve data from Bookmark-> "+err.Error())
	}

	return &v1.ListBookmarksRes{
		Api:       apiVersion,
		Bookmarks: bookmarks,
	}, nil
}

func (s *BookmarkServiceServer) DeleteBookmark(ctx context.Context, req *v1.DeleteBookmarkReq) (*v1.DeleteBookmarkRes, error) {
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}

	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}

	var query, param string

	if req.Bookmark.Id != 0 {
		query = "id = ?"
		param = strconv.Itoa(int(req.Bookmark.Id))
	}

	if req.Bookmark.Url != "" {
		query = "url = ?"
		param = req.Bookmark.Url
	}

	if query == "" || param == "" {
		return nil, status.Error(codes.InvalidArgument, "failed to insert into todo")
	}

	squery := fmt.Sprintf("DELETE FROM bookmarks where %s", query)

	rows, err := c.QueryContext(ctx, squery, param)
	defer c.Close()

	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to delete from Bookmarks "+err.Error())
	}
	defer rows.Close()

	bookmarks := []*v1.Bookmark{}
	for rows.Next() {
		bk := new(v1.Bookmark)
		if err := rows.Scan(&bk.Id, &bk.Url); err != nil {
			return nil, status.Error(codes.Unknown, "failed to retrieve field values from Bookmark row-> "+err.Error())
		}

		bookmarks = append(bookmarks, bk)
	}

	if err := rows.Err(); err != nil {
		return nil, status.Error(codes.Unknown, "failed to retrieve data from Bookmark-> "+err.Error())
	}

	return &v1.DeleteBookmarkRes{
		Api:       apiVersion,
		Bookmarks: bookmarks,
	}, nil
}

func (s *BookmarkServiceServer) ReadBookmark(ctx context.Context, req *v1.ReadBookmarkReq) (*v1.ReadBookmarkRes, error) {
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}

	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}

	if req.Bookmark.Url == "" {
		return nil, status.Error(codes.InvalidArgument, "failed to insert into todo")
	}

	param := "'%" + req.Bookmark.Url + "%'"
	query := fmt.Sprintf("select id, url from bookmarks where url LIKE %s", param)
	rows, err := c.QueryContext(ctx, query)
	defer c.Close()

	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to retrieve from Bookmarks "+err.Error())
	}
	defer rows.Close()

	bookmarks := []*v1.Bookmark{}
	for rows.Next() {
		bk := new(v1.Bookmark)
		if err := rows.Scan(&bk.Id, &bk.Url); err != nil {
			return nil, status.Error(codes.Unknown, "failed to retrieve field values from Bookmark row-> "+err.Error())
		}

		bookmarks = append(bookmarks, bk)
	}

	if err := rows.Err(); err != nil {
		return nil, status.Error(codes.Unknown, "failed to retrieve data from Bookmark-> "+err.Error())
	}

	return &v1.ReadBookmarkRes{
		Api:       apiVersion,
		Bookmarks: bookmarks,
	}, nil
}

func (s *BookmarkServiceServer) mustEmbedUnimplementedBookmarkServiceServer() {}

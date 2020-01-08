package memory

import (
	

	"github.com/slim-crown/issue-1-REST/pkg/domain/post"
)

//postRepository ...
type postRepository struct {
	cache         map[int]post.Post
	secondaryRepo *post.Repository
}

// NewPostRepository returns a struct that implements the post.Repository using
func NewPostRepository(secondaryRepo *post.Repository) post.Repository {
	return &postRepository{cache: make(map[int]post.Post), secondaryRepo: secondaryRepo}
}

// cachePost is just a helper function to update the cache with new states of the struct
func (repo *postRepository) cachePost(id int) error {
	u, err := (*repo.secondaryRepo).GetPost(id)
	if err != nil {
		return err
	}
	repo.cache[id] = *u
	return nil
}

// GetPost gets the Post stored under the given id.
func (repo *postRepository) GetPost(id int) (*post.Post, error) {
	if _, ok := repo.cache[id]; ok == false {
		r, err := (*repo.secondaryRepo).GetPost(id)
		if err != nil {
			return nil, err
		}
		repo.cache[id] = *r
	}
	r := repo.cache[id]
	return &r, nil
	
}

// DeletePost Deletes the Post stored under the given id.
func (repo *postRepository) DeletePost(id int) error {
	err:=(*repo.secondaryRepo).DeletePost(id)
	if err==nil{
		delete(repo.cache,id)
	}
	return err
}

// AddPost Adds the Post stored under the given id.
func (repo *postRepository) AddPost(p *post.Post) (*post.Post, error) {
	p,err:=(*repo.secondaryRepo).AddPost(p)
	if(err==nil){
		repo.cache[p.ID]=*p
	}
	return p,err
}

//UpdatePost updates the post with given id and post struct
func (repo *postRepository) UpdatePost(pos *post.Post, id int) (*post.Post, error) {
	p,err:=(*repo.secondaryRepo).UpdatePost(pos, id)
	if(err==nil){
		repo.cache[p.ID]=*p
	}
	return p,err
}

// GetPostReleases(p *Post) ([]*Release, error)
// GetPostRelease(pId int, rId int) (*Release, error)

func (repo *postRepository) SearchPost(pattern string, by post.SortBy, order post.SortOrder, limit int, offset int) ([]*post.Post, error) {
	pos,err:=(*repo.secondaryRepo).SearchPost(pattern, by, order, limit, offset)
	if err==nil{
		for _,p:= range pos{
			repo.cache[p.ID]=*p

		}
	}
	return pos,err
}
func (repo *postRepository) GetPostStar(id int, username string) (*post.Star, error) {
	return (*repo.secondaryRepo).GetPostStar(id, username)
}
func (repo *postRepository) DeletePostStar(id int, username string) error {
	err:= (*repo.secondaryRepo).DeletePostStar(id, username)
	if err==nil{
		errs:=repo.cachePost(id)
		if errs!=nil{
			return errs
		}
		
	}
	return err
}
func (repo *postRepository) AddPostStar(id int, star *post.Star) (*post.Star, error) {
	s,err:=(*repo.secondaryRepo).AddPostStar(id, star)
	if err==nil{
		errs:=repo.cachePost(id)
		if errs!=nil{
			return s,errs
		}
	}
	return s,err
}
func (repo *postRepository) UpdatePostStar(id int, star *post.Star) (*post.Star, error) {
	s,err:=(*repo.secondaryRepo).UpdatePostStar(id, star)
	if err==nil{
		errs:=repo.cachePost(id)
		if errs!=nil{
			return s,errs
		}
	}
	return s,err
}

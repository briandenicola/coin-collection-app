package handlers

import (
	"net/http"
	"strconv"

	"github.com/briandenicola/ancient-coins-api/database"
	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
)

type SocialHandler struct{}

func NewSocialHandler() *SocialHandler {
	return &SocialHandler{}
}

// FollowUser sends a follow request to another user.
func (h *SocialHandler) FollowUser(c *gin.Context) {
	logger := services.AppLogger
	userID := c.GetUint("userId")
	targetID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if uint(targetID) == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot follow yourself"})
		return
	}

	// Verify target exists and is public
	var target models.User
	if err := database.DB.First(&target, targetID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	if !target.IsPublic {
		c.JSON(http.StatusForbidden, gin.H{"error": "This user is not accepting followers"})
		return
	}

	// Check if blocked
	var existing models.Follow
	if err := database.DB.Where("follower_id = ? AND following_id = ?", userID, targetID).First(&existing).Error; err == nil {
		if existing.Status == "blocked" {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are blocked by this user"})
			return
		}
		if existing.Status == "pending" {
			c.JSON(http.StatusConflict, gin.H{"error": "Follow request already pending"})
			return
		}
		c.JSON(http.StatusConflict, gin.H{"error": "Already following this user"})
		return
	}

	follow := models.Follow{
		FollowerID:  userID,
		FollowingID: uint(targetID),
		Status:      "pending",
	}

	if err := database.DB.Create(&follow).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Already following this user"})
		return
	}

	logger.Info("social", "User %d sent follow request to user %d", userID, targetID)
	c.JSON(http.StatusCreated, gin.H{"message": "Follow request sent"})
}

// UnfollowUser unfollows a user (removes any follow relationship).
func (h *SocialHandler) UnfollowUser(c *gin.Context) {
	logger := services.AppLogger
	userID := c.GetUint("userId")
	targetID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	result := database.DB.Where("follower_id = ? AND following_id = ? AND status != ?", userID, targetID, "blocked").Delete(&models.Follow{})
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not following this user"})
		return
	}

	logger.Info("social", "User %d unfollowed user %d", userID, targetID)
	c.JSON(http.StatusOK, gin.H{"message": "Unfollowed user"})
}

// AcceptFollower accepts a pending follow request.
func (h *SocialHandler) AcceptFollower(c *gin.Context) {
	logger := services.AppLogger
	userID := c.GetUint("userId")
	followerID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	result := database.DB.Model(&models.Follow{}).
		Where("follower_id = ? AND following_id = ? AND status = ?", followerID, userID, "pending").
		Update("status", "accepted")
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No pending follow request from this user"})
		return
	}

	logger.Info("social", "User %d accepted follower %d", userID, followerID)
	c.JSON(http.StatusOK, gin.H{"message": "Follower accepted"})
}

// BlockFollower blocks a user from following.
func (h *SocialHandler) BlockFollower(c *gin.Context) {
	logger := services.AppLogger
	userID := c.GetUint("userId")
	followerID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Check if there's an existing follow record
	var follow models.Follow
	if err := database.DB.Where("follower_id = ? AND following_id = ?", followerID, userID).First(&follow).Error; err != nil {
		// No existing relationship — create a blocked record
		follow = models.Follow{
			FollowerID:  uint(followerID),
			FollowingID: userID,
			Status:      "blocked",
		}
		database.DB.Create(&follow)
	} else {
		database.DB.Model(&follow).Update("status", "blocked")
	}

	logger.Info("social", "User %d blocked follower %d", userID, followerID)
	c.JSON(http.StatusOK, gin.H{"message": "User blocked"})
}

// UnblockFollower removes a block on a user.
func (h *SocialHandler) UnblockFollower(c *gin.Context) {
	logger := services.AppLogger
	userID := c.GetUint("userId")
	followerID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	result := database.DB.Where("follower_id = ? AND following_id = ? AND status = ?", followerID, userID, "blocked").Delete(&models.Follow{})
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User is not blocked"})
		return
	}

	logger.Info("social", "User %d unblocked user %d", userID, followerID)
	c.JSON(http.StatusOK, gin.H{"message": "User unblocked"})
}

// GetBlockedUsers returns users blocked by the authenticated user.
func (h *SocialHandler) GetBlockedUsers(c *gin.Context) {
	userID := c.GetUint("userId")

	var follows []models.Follow
	database.DB.Where("following_id = ? AND status = ?", userID, "blocked").Find(&follows)

	blockedIDs := make([]uint, len(follows))
	for i, f := range follows {
		blockedIDs[i] = f.FollowerID
	}

	if len(blockedIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{"blocked": []interface{}{}})
		return
	}

	var users []models.User
	database.DB.Where("id IN ?", blockedIDs).Find(&users)

	var result []gin.H
	for _, u := range users {
		result = append(result, gin.H{
			"id":         u.ID,
			"username":   u.Username,
			"avatarPath": u.AvatarPath,
		})
	}
	if result == nil {
		result = []gin.H{}
	}
	c.JSON(http.StatusOK, gin.H{"blocked": result})
}

// GetFollowers returns users who follow the authenticated user (pending + accepted).
func (h *SocialHandler) GetFollowers(c *gin.Context) {
	userID := c.GetUint("userId")

	var follows []models.Follow
	database.DB.Where("following_id = ? AND status IN ?", userID, []string{"pending", "accepted"}).Find(&follows)

	followerIDs := make([]uint, len(follows))
	statusMap := map[uint]string{}
	for i, f := range follows {
		followerIDs[i] = f.FollowerID
		statusMap[f.FollowerID] = f.Status
	}

	if len(followerIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{"followers": []interface{}{}})
		return
	}

	var users []models.User
	database.DB.Where("id IN ?", followerIDs).Find(&users)

	var result []gin.H
	for _, u := range users {
		result = append(result, gin.H{
			"id":         u.ID,
			"username":   u.Username,
			"avatarPath": u.AvatarPath,
			"isPublic":   u.IsPublic,
			"bio":        u.Bio,
			"status":     statusMap[u.ID],
		})
	}
	if result == nil {
		result = []gin.H{}
	}

	c.JSON(http.StatusOK, gin.H{"followers": result})
}

// GetFollowing returns users the authenticated user follows (accepted only).
func (h *SocialHandler) GetFollowing(c *gin.Context) {
	userID := c.GetUint("userId")

	var follows []models.Follow
	database.DB.Where("follower_id = ? AND status = ?", userID, "accepted").Find(&follows)

	followingIDs := make([]uint, len(follows))
	for i, f := range follows {
		followingIDs[i] = f.FollowingID
	}

	if len(followingIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{"following": []interface{}{}})
		return
	}

	var users []models.User
	database.DB.Where("id IN ?", followingIDs).Find(&users)

	result := buildUserList(users, userID)
	c.JSON(http.StatusOK, gin.H{"following": result})
}

// SearchUsers searches for users by username prefix (public users only).
func (h *SocialHandler) SearchUsers(c *gin.Context) {
	userID := c.GetUint("userId")
	query := c.Query("q")
	if query == "" || len(query) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query must be at least 2 characters"})
		return
	}

	var users []models.User
	database.DB.Where("username LIKE ? AND id != ? AND is_public = ?", query+"%", userID, true).Limit(20).Find(&users)

	// Get follow status for each result
	var follows []models.Follow
	database.DB.Where("follower_id = ?", userID).Find(&follows)
	followStatusMap := map[uint]string{}
	for _, f := range follows {
		followStatusMap[f.FollowingID] = f.Status
	}

	var result []gin.H
	for _, u := range users {
		var coinCount int64
		database.DB.Model(&models.Coin{}).Where("user_id = ? AND is_wishlist = false AND is_sold = false AND is_private = false", u.ID).Count(&coinCount)
		status := followStatusMap[u.ID] // "" if not following
		result = append(result, gin.H{
			"id":           u.ID,
			"username":     u.Username,
			"avatarPath":   u.AvatarPath,
			"isPublic":     u.IsPublic,
			"bio":          u.Bio,
			"isFollowing":  status == "accepted",
			"followStatus": status,
			"coinCount":    coinCount,
		})
	}

	if result == nil {
		result = []gin.H{}
	}
	c.JSON(http.StatusOK, gin.H{"users": result})
}

// GetPublicProfile returns a user's public profile.
func (h *SocialHandler) GetPublicProfile(c *gin.Context) {
	username := c.Param("username")
	currentUserID := c.GetUint("userId")

	var user models.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Check follow status
	var followStatus string
	var isFollowing bool
	var follow models.Follow
	if err := database.DB.Where("follower_id = ? AND following_id = ?", currentUserID, user.ID).First(&follow).Error; err == nil {
		followStatus = follow.Status
		isFollowing = follow.Status == "accepted"
	}

	var coinCount int64
	if user.IsPublic && isFollowing {
		database.DB.Model(&models.Coin{}).Where("user_id = ? AND is_wishlist = false AND is_sold = false AND is_private = false", user.ID).Count(&coinCount)
	}

	var followerCount, followingCount int64
	database.DB.Model(&models.Follow{}).Where("following_id = ? AND status = ?", user.ID, "accepted").Count(&followerCount)
	database.DB.Model(&models.Follow{}).Where("follower_id = ? AND status = ?", user.ID, "accepted").Count(&followingCount)

	c.JSON(http.StatusOK, gin.H{
		"id":             user.ID,
		"username":       user.Username,
		"avatarPath":     user.AvatarPath,
		"isPublic":       user.IsPublic,
		"bio":            user.Bio,
		"isFollowing":    isFollowing,
		"followStatus":   followStatus,
		"coinCount":      coinCount,
		"followerCount":  followerCount,
		"followingCount": followingCount,
	})
}

// GetFollowingCoins returns a followed user's public coins (limited fields for gallery).
func (h *SocialHandler) GetFollowingCoins(c *gin.Context) {
	currentUserID := c.GetUint("userId")
	targetID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Verify accepted following relationship
	var follow models.Follow
	if err := database.DB.Where("follower_id = ? AND following_id = ? AND status = ?", currentUserID, targetID, "accepted").First(&follow).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "You must be an accepted follower to view their coins"})
		return
	}

	// Verify target is public
	var target models.User
	if err := database.DB.First(&target, targetID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	if !target.IsPublic {
		c.JSON(http.StatusForbidden, gin.H{"error": "This user's collection is private"})
		return
	}

	var coins []models.Coin
	database.DB.Where("user_id = ? AND is_wishlist = false AND is_sold = false AND is_private = false", targetID).
		Preload("Images").
		Order("updated_at DESC").
		Find(&coins)

	var result []gin.H
	for _, coin := range coins {
		result = append(result, limitedCoinData(coin))
	}
	if result == nil {
		result = []gin.H{}
	}

	c.JSON(http.StatusOK, gin.H{
		"coins":    result,
		"username": target.Username,
	})
}

// GetFollowingCoinDetail returns a single coin with limited fields for a follower.
func (h *SocialHandler) GetFollowingCoinDetail(c *gin.Context) {
	currentUserID := c.GetUint("userId")
	targetID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	coinID, err := strconv.ParseUint(c.Param("coinId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coin ID"})
		return
	}

	// Verify accepted following
	var follow models.Follow
	if err := database.DB.Where("follower_id = ? AND following_id = ? AND status = ?", currentUserID, targetID, "accepted").First(&follow).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "You must be an accepted follower to view their coins"})
		return
	}

	var coin models.Coin
	if err := database.DB.Where("id = ? AND user_id = ? AND is_private = false AND is_wishlist = false AND is_sold = false", coinID, targetID).
		Preload("Images").First(&coin).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Coin not found"})
		return
	}

	// Get comments
	var comments []models.CoinComment
	database.DB.Where("coin_id = ?", coinID).Order("created_at DESC").Find(&comments)

	var commentResults []gin.H
	for _, cm := range comments {
		var commenter models.User
		database.DB.First(&commenter, cm.UserID)
		commentResults = append(commentResults, gin.H{
			"id":         cm.ID,
			"coinId":     cm.CoinID,
			"userId":     cm.UserID,
			"username":   commenter.Username,
			"avatarPath": commenter.AvatarPath,
			"comment":    cm.Comment,
			"rating":     cm.Rating,
			"createdAt":  cm.CreatedAt,
		})
	}
	if commentResults == nil {
		commentResults = []gin.H{}
	}

	// Get aggregate rating
	type RatingResult struct {
		Avg   float64
		Count int64
	}
	var ratingResult RatingResult
	database.DB.Model(&models.CoinComment{}).
		Select("COALESCE(AVG(NULLIF(rating, 0)), 0) as avg, COUNT(NULLIF(rating, 0)) as count").
		Where("coin_id = ?", coinID).
		Scan(&ratingResult)

	// Get current user's rating
	var userRating int
	var userComment models.CoinComment
	if err := database.DB.Where("coin_id = ? AND user_id = ? AND rating > 0", coinID, currentUserID).First(&userComment).Error; err == nil {
		userRating = userComment.Rating
	}

	result := limitedCoinData(coin)
	result["comments"] = commentResults
	result["rating"] = gin.H{
		"average":    ratingResult.Avg,
		"count":      ratingResult.Count,
		"userRating": userRating,
	}

	c.JSON(http.StatusOK, result)
}

// AddComment adds a comment (with optional rating) to a coin.
func (h *SocialHandler) AddComment(c *gin.Context) {
	logger := services.AppLogger
	currentUserID := c.GetUint("userId")
	coinID, err := strconv.ParseUint(c.Param("coinId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coin ID"})
		return
	}

	var req struct {
		Comment string `json:"comment" binding:"required"`
		Rating  int    `json:"rating"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Rating < 0 || req.Rating > 5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Rating must be between 0 and 5"})
		return
	}

	var coin models.Coin
	if err := database.DB.First(&coin, coinID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Coin not found"})
		return
	}

	if coin.IsPrivate {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot comment on a private coin"})
		return
	}

	// Verify accepted follower relationship (unless coin owner)
	if coin.UserID != currentUserID {
		var follow models.Follow
		if err := database.DB.Where("follower_id = ? AND following_id = ? AND status = ?", currentUserID, coin.UserID, "accepted").First(&follow).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "You must be an accepted follower to comment on their coins"})
			return
		}
	}

	comment := models.CoinComment{
		CoinID:  uint(coinID),
		UserID:  currentUserID,
		Comment: req.Comment,
		Rating:  req.Rating,
	}

	if err := database.DB.Create(&comment).Error; err != nil {
		logger.Error("social", "Failed to create comment: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create comment"})
		return
	}

	var user models.User
	database.DB.First(&user, currentUserID)

	logger.Info("social", "User %d commented on coin %d", currentUserID, coinID)
	c.JSON(http.StatusCreated, gin.H{
		"id":         comment.ID,
		"coinId":     comment.CoinID,
		"userId":     comment.UserID,
		"username":   user.Username,
		"avatarPath": user.AvatarPath,
		"comment":    comment.Comment,
		"rating":     comment.Rating,
		"createdAt":  comment.CreatedAt,
	})
}

// GetComments returns comments on a coin.
func (h *SocialHandler) GetComments(c *gin.Context) {
	currentUserID := c.GetUint("userId")
	coinID, err := strconv.ParseUint(c.Param("coinId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coin ID"})
		return
	}

	var coin models.Coin
	if err := database.DB.First(&coin, coinID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Coin not found"})
		return
	}

	// Must be owner or accepted follower
	if coin.UserID != currentUserID {
		var follow models.Follow
		if err := database.DB.Where("follower_id = ? AND following_id = ? AND status = ?", currentUserID, coin.UserID, "accepted").First(&follow).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
	}

	var comments []models.CoinComment
	database.DB.Where("coin_id = ?", coinID).Order("created_at DESC").Find(&comments)

	var result []gin.H
	for _, cm := range comments {
		var commenter models.User
		database.DB.First(&commenter, cm.UserID)
		result = append(result, gin.H{
			"id":         cm.ID,
			"coinId":     cm.CoinID,
			"userId":     cm.UserID,
			"username":   commenter.Username,
			"avatarPath": commenter.AvatarPath,
			"comment":    cm.Comment,
			"rating":     cm.Rating,
			"createdAt":  cm.CreatedAt,
		})
	}
	if result == nil {
		result = []gin.H{}
	}

	c.JSON(http.StatusOK, gin.H{"comments": result})
}

// DeleteComment deletes a comment. Owner of coin or commenter can delete.
func (h *SocialHandler) DeleteComment(c *gin.Context) {
	logger := services.AppLogger
	currentUserID := c.GetUint("userId")
	coinID, err := strconv.ParseUint(c.Param("coinId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coin ID"})
		return
	}
	commentID, err := strconv.ParseUint(c.Param("commentId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment ID"})
		return
	}

	var comment models.CoinComment
	if err := database.DB.Where("id = ? AND coin_id = ?", commentID, coinID).First(&comment).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		return
	}

	var coin models.Coin
	database.DB.First(&coin, coinID)
	if comment.UserID != currentUserID && coin.UserID != currentUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete this comment"})
		return
	}

	database.DB.Delete(&comment)
	logger.Info("social", "Comment %d deleted by user %d", commentID, currentUserID)
	c.JSON(http.StatusOK, gin.H{"message": "Comment deleted"})
}

// RateCoin upserts a star rating for a coin.
func (h *SocialHandler) RateCoin(c *gin.Context) {
	currentUserID := c.GetUint("userId")
	coinID, err := strconv.ParseUint(c.Param("coinId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coin ID"})
		return
	}

	var req struct {
		Rating int `json:"rating" binding:"required,min=1,max=5"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Rating must be between 1 and 5"})
		return
	}

	var coin models.Coin
	if err := database.DB.First(&coin, coinID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Coin not found"})
		return
	}

	// Must be accepted follower (unless owner)
	if coin.UserID != currentUserID {
		var follow models.Follow
		if err := database.DB.Where("follower_id = ? AND following_id = ? AND status = ?", currentUserID, coin.UserID, "accepted").First(&follow).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "You must be an accepted follower to rate their coins"})
			return
		}
	}

	var existing models.CoinComment
	if err := database.DB.Where("coin_id = ? AND user_id = ? AND rating > 0", coinID, currentUserID).First(&existing).Error; err == nil {
		database.DB.Model(&existing).Update("rating", req.Rating)
	} else {
		rating := models.CoinComment{
			CoinID:  uint(coinID),
			UserID:  currentUserID,
			Comment: "",
			Rating:  req.Rating,
		}
		database.DB.Create(&rating)
	}

	type RatingResult struct {
		Avg   float64
		Count int64
	}
	var ratingResult RatingResult
	database.DB.Model(&models.CoinComment{}).
		Select("COALESCE(AVG(NULLIF(rating, 0)), 0) as avg, COUNT(NULLIF(rating, 0)) as count").
		Where("coin_id = ?", coinID).
		Scan(&ratingResult)

	c.JSON(http.StatusOK, gin.H{
		"average":    ratingResult.Avg,
		"count":      ratingResult.Count,
		"userRating": req.Rating,
	})
}

// GetCoinRating returns aggregate and user-specific rating for a coin.
func (h *SocialHandler) GetCoinRating(c *gin.Context) {
	currentUserID := c.GetUint("userId")
	coinID, err := strconv.ParseUint(c.Param("coinId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coin ID"})
		return
	}

	type RatingResult struct {
		Avg   float64
		Count int64
	}
	var ratingResult RatingResult
	database.DB.Model(&models.CoinComment{}).
		Select("COALESCE(AVG(NULLIF(rating, 0)), 0) as avg, COUNT(NULLIF(rating, 0)) as count").
		Where("coin_id = ?", coinID).
		Scan(&ratingResult)

	var userRating int
	var userComment models.CoinComment
	if err := database.DB.Where("coin_id = ? AND user_id = ? AND rating > 0", coinID, currentUserID).First(&userComment).Error; err == nil {
		userRating = userComment.Rating
	}

	c.JSON(http.StatusOK, gin.H{
		"average":    ratingResult.Avg,
		"count":      ratingResult.Count,
		"userRating": userRating,
	})
}

// Helper: build user list with coin counts
func buildUserList(users []models.User, currentUserID uint) []gin.H {
	var follows []models.Follow
	database.DB.Where("follower_id = ?", currentUserID).Find(&follows)
	followSet := map[uint]bool{}
	for _, f := range follows {
		if f.Status == "accepted" {
			followSet[f.FollowingID] = true
		}
	}

	var result []gin.H
	for _, u := range users {
		var coinCount int64
		if u.IsPublic {
			database.DB.Model(&models.Coin{}).Where("user_id = ? AND is_wishlist = false AND is_sold = false AND is_private = false", u.ID).Count(&coinCount)
		}
		result = append(result, gin.H{
			"id":          u.ID,
			"username":    u.Username,
			"avatarPath":  u.AvatarPath,
			"isPublic":    u.IsPublic,
			"bio":         u.Bio,
			"isFollowing": followSet[u.ID],
			"coinCount":   coinCount,
		})
	}
	if result == nil {
		result = []gin.H{}
	}
	return result
}

// Helper: strip sensitive fields from coin for follower view
func limitedCoinData(coin models.Coin) gin.H {
	return gin.H{
		"id":           coin.ID,
		"name":         coin.Name,
		"category":     coin.Category,
		"denomination": coin.Denomination,
		"ruler":        coin.Ruler,
		"era":          coin.Era,
		"material":     coin.Material,
		"grade":        coin.Grade,
		"images":       coin.Images,
	}
}

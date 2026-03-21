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

// FollowUser follows another user.
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

	// Verify target exists
	var target models.User
	if err := database.DB.First(&target, targetID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	follow := models.Follow{
		FollowerID:  userID,
		FollowingID: uint(targetID),
	}

	if err := database.DB.Create(&follow).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Already following this user"})
		return
	}

	logger.Info("social", "User %d followed user %d", userID, targetID)
	c.JSON(http.StatusCreated, gin.H{"message": "Now following user"})
}

// UnfollowUser unfollows a user.
func (h *SocialHandler) UnfollowUser(c *gin.Context) {
	logger := services.AppLogger
	userID := c.GetUint("userId")
	targetID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	result := database.DB.Where("follower_id = ? AND following_id = ?", userID, targetID).Delete(&models.Follow{})
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not following this user"})
		return
	}

	logger.Info("social", "User %d unfollowed user %d", userID, targetID)
	c.JSON(http.StatusOK, gin.H{"message": "Unfollowed user"})
}

// GetFollowers returns users who follow the authenticated user.
func (h *SocialHandler) GetFollowers(c *gin.Context) {
	userID := c.GetUint("userId")

	var follows []models.Follow
	database.DB.Where("following_id = ?", userID).Find(&follows)

	followerIDs := make([]uint, len(follows))
	for i, f := range follows {
		followerIDs[i] = f.FollowerID
	}

	if len(followerIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{"followers": []interface{}{}})
		return
	}

	var users []models.User
	database.DB.Where("id IN ?", followerIDs).Find(&users)

	result := buildUserList(users, userID)
	c.JSON(http.StatusOK, gin.H{"followers": result})
}

// GetFollowing returns users the authenticated user follows.
func (h *SocialHandler) GetFollowing(c *gin.Context) {
	userID := c.GetUint("userId")

	var follows []models.Follow
	database.DB.Where("follower_id = ?", userID).Find(&follows)

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

// SearchUsers searches for users by username prefix.
func (h *SocialHandler) SearchUsers(c *gin.Context) {
	userID := c.GetUint("userId")
	query := c.Query("q")
	if query == "" || len(query) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query must be at least 2 characters"})
		return
	}

	var users []models.User
	database.DB.Where("username LIKE ? AND id != ?", query+"%", userID).Limit(20).Find(&users)

	// Get follow status for each result
	var followedIDs []uint
	var follows []models.Follow
	database.DB.Where("follower_id = ?", userID).Find(&follows)
	for _, f := range follows {
		followedIDs = append(followedIDs, f.FollowingID)
	}

	followSet := map[uint]bool{}
	for _, id := range followedIDs {
		followSet[id] = true
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

	// Check if current user follows this user
	var isFollowing bool
	var follow models.Follow
	if err := database.DB.Where("follower_id = ? AND following_id = ?", currentUserID, user.ID).First(&follow).Error; err == nil {
		isFollowing = true
	}

	var coinCount int64
	if user.IsPublic || isFollowing {
		database.DB.Model(&models.Coin{}).Where("user_id = ? AND is_wishlist = false AND is_sold = false AND is_private = false", user.ID).Count(&coinCount)
	}

	var followerCount, followingCount int64
	database.DB.Model(&models.Follow{}).Where("following_id = ?", user.ID).Count(&followerCount)
	database.DB.Model(&models.Follow{}).Where("follower_id = ?", user.ID).Count(&followingCount)

	c.JSON(http.StatusOK, gin.H{
		"id":             user.ID,
		"username":       user.Username,
		"avatarPath":     user.AvatarPath,
		"isPublic":       user.IsPublic,
		"bio":            user.Bio,
		"isFollowing":    isFollowing,
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

	// Verify following relationship
	var follow models.Follow
	if err := database.DB.Where("follower_id = ? AND following_id = ?", currentUserID, targetID).First(&follow).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "You must follow this user to view their coins"})
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

	// Strip sensitive fields
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

	// Verify following
	var follow models.Follow
	if err := database.DB.Where("follower_id = ? AND following_id = ?", currentUserID, targetID).First(&follow).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "You must follow this user to view their coins"})
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

	// Enrich comments with usernames and avatars
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

	// Get coin and verify access
	var coin models.Coin
	if err := database.DB.First(&coin, coinID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Coin not found"})
		return
	}

	if coin.IsPrivate {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot comment on a private coin"})
		return
	}

	// Verify follower relationship (unless it's the coin owner)
	if coin.UserID != currentUserID {
		var follow models.Follow
		if err := database.DB.Where("follower_id = ? AND following_id = ?", currentUserID, coin.UserID).First(&follow).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "You must follow this user to comment on their coins"})
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

	// Return enriched comment
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

	// Must be owner or follower
	if coin.UserID != currentUserID {
		var follow models.Follow
		if err := database.DB.Where("follower_id = ? AND following_id = ?", currentUserID, coin.UserID).First(&follow).Error; err != nil {
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

	// Check permission: commenter or coin owner
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

	// Must follow coin owner (unless owner themselves)
	if coin.UserID != currentUserID {
		var follow models.Follow
		if err := database.DB.Where("follower_id = ? AND following_id = ?", currentUserID, coin.UserID).First(&follow).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "You must follow this user to rate their coins"})
			return
		}
	}

	// Upsert: find existing rating comment or create a rating-only record
	var existing models.CoinComment
	if err := database.DB.Where("coin_id = ? AND user_id = ? AND rating > 0", coinID, currentUserID).First(&existing).Error; err == nil {
		database.DB.Model(&existing).Update("rating", req.Rating)
	} else {
		// Create a rating-only comment (empty comment text is allowed for pure ratings)
		rating := models.CoinComment{
			CoinID:  uint(coinID),
			UserID:  currentUserID,
			Comment: "",
			Rating:  req.Rating,
		}
		database.DB.Create(&rating)
	}

	// Return aggregate
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
	// Get IDs of who the current user follows
	var follows []models.Follow
	database.DB.Where("follower_id = ?", currentUserID).Find(&follows)
	followSet := map[uint]bool{}
	for _, f := range follows {
		followSet[f.FollowingID] = true
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

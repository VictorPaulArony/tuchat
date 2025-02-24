class SocialApp {
    constructor() {
      this.initializeDOM();
      this.posts = [];
      this.setupEventListeners();
      this.generateRealisticPosts();
      this.setupFilterMechanism();
    }
  
    initializeDOM() {
      this.domElements = {
        postsContainer: document.querySelector('#posts-container'),
        createPostModal: document.querySelector('.create-post-modal'),
        createPostTriggers: document.querySelectorAll('.create-post-trigger'),
        sharePostBtn: document.querySelector('.create-post-modal .btn-primary'),
        postTextarea: document.querySelector('.create-post-modal textarea'),
        categoryCheckboxes: document.querySelectorAll('.create-post-modal .category-checkboxes input')
      };
    }
  
    setupEventListeners() {
      this.setupCreatePostListeners();
      this.setupPostInteractionListeners();
      this.setupFilterListeners();
    }
  
    setupCreatePostListeners() {
      const { createPostTriggers, createPostModal, sharePostBtn, postTextarea, categoryCheckboxes } = this.domElements;
  
      createPostTriggers.forEach(trigger => {
        trigger.addEventListener('click', () => createPostModal.style.display = 'flex');
      });
  
      sharePostBtn.addEventListener('click', () => {
        const postContent = postTextarea.value.trim();
        const selectedCategories = Array.from(categoryCheckboxes)
          .filter(checkbox => checkbox.checked)
          .map(checkbox => checkbox.value);
  
        if (postContent) {
          this.createPost(postContent, selectedCategories);
          this.resetPostModalForm();
        }
      });
    }
  
    resetPostModalForm() {
      const { postTextarea, categoryCheckboxes, createPostModal } = this.domElements;
      postTextarea.value = '';
      categoryCheckboxes.forEach(checkbox => checkbox.checked = false);
      createPostModal.style.display = 'none';
    }
  
    setupPostInteractionListeners() {
      this.setupEventDelegation();
    }
  
    setupEventDelegation() {
      this.domElements.postsContainer.addEventListener('click', (event) => {
        const target = event.target;
        const postElement = target.closest('.post');
  
        if (postElement) {
          const postId = postElement.dataset.postId;
          const post = this.posts.find(p => p.id == postId);
  
          if (target.closest('.like-btn')) this.toggleLike(postElement);
          if (target.closest('.dislike-btn')) this.toggleDislike(postElement);
          if (target.closest('.comments-toggle')) this.toggleComments(postElement);
          if (target.closest('.send-comment')) this.handleCommentSubmission(postElement);
          if (target.closest('.toggle-likes-details')) this.toggleLikesDetails(postElement);
        }
      });
    }
  
    setupFilterListeners() {
      const filterButtons = document.querySelectorAll('.filter-btn');
      const categoriesDropdown = document.querySelector('.categories-dropdown');
      const categorySelect = document.getElementById('category-select');
  
      filterButtons.forEach(button => {
        button.addEventListener('click', () => {
          // Remove active class from all buttons
          filterButtons.forEach(btn => btn.classList.remove('active'));
          button.classList.add('active');
  
          // Hide/show categories dropdown
          categoriesDropdown.classList.toggle('hidden', button.dataset.filter !== 'categories');
  
          // Apply filtering
          this.filterPosts(button.dataset.filter, categorySelect.value);
        });
      });
  
      // Add event listener for category selection
      categorySelect.addEventListener('change', () => {
        const activeFilter = document.querySelector('.filter-btn.active').dataset.filter;
        this.filterPosts(activeFilter, categorySelect.value);
      });
    }
  
    generateRealisticPosts() {
      const realisticPosts = [
        { 
          id: 1,
          user: 'Adventure Seeker',
          userAvatar: 'https://randomuser.me/api/portraits/women/44.jpg',
          image: 'https://images.unsplash.com/photo-1519671482749-fd09be7ccfb4?ixlib=rb-4.0.3&auto=format&fit=crop&w=1470&q=80', 
          caption: 'Sunrise at the mountain peak. Nothing beats this view! ',
          categories: ['travel', 'photography'], 
          likes: 1256, 
          dislikes: 23,
          comments: [
            { user: 'Travel Buddy', text: 'Absolutely breathtaking!' },
            { user: 'Nature Lover', text: 'Wish I was there right now' }
          ],
          likedBy: ['Emma', 'Jack', 'Sophia'],
          dislikedBy: ['Mike']
        },
        { 
          id: 2,
          user: 'Culinary Explorer',
          userAvatar: 'https://randomuser.me/api/portraits/men/32.jpg',
          image: 'https://images.unsplash.com/photo-1555939594-58639d380623?ixlib=rb-4.0.3&auto=format&fit=crop&w=1471&q=80', 
          caption: 'Today\'s gourmet creation - homemade pasta with truffle sauce ',
          categories: ['food', 'lifestyle'], 
          likes: 789, 
          dislikes: 12,
          comments: [
            { user: 'Foodie', text: 'This looks delicious!' },
            { user: 'Chef', text: 'Amazing plating skills' }
          ],
          likedBy: ['David', 'Rachel', 'Chris'],
          dislikedBy: ['Lauren']
        },
        { 
          id: 3,
          user: 'Urban Photographer',
          userAvatar: 'https://randomuser.me/api/portraits/women/68.jpg',
          image: 'https://images.unsplash.com/photo-1516450360452-9312f5e86fc7?ixlib=rb-4.0.3&auto=format&fit=crop&w=1470&q=80', 
          caption: 'City lights never looked so magical ',
          categories: ['photography', 'travel'], 
          likes: 2345, 
          dislikes: 45,
          comments: [
            { user: 'Art Lover', text: 'Incredible composition!' },
            { user: 'Traveler', text: 'Which city is this?' }
          ],
          likedBy: ['Alex', 'Olivia', 'Ethan'],
          dislikedBy: ['Zoe']
        }
      ];
  
      realisticPosts.forEach(post => {
        this.posts.push(post);
        const postElement = this.createPostElement(post);
        this.domElements.postsContainer.appendChild(postElement);
      });
    }
  
    createPostElement(post) {
      const postElement = document.createElement('div');
      postElement.classList.add('post');
      postElement.dataset.postId = post.id;
      
      // Add category tags to post element
      const categoryTags = post.categories ? post.categories.map(category => 
        `<span class="category-tag">${category}</span>`
      ).join('') : '';
  
      postElement.innerHTML = `
        <div class="post-header">
          <img src="${post.userAvatar}" class="user-avatar rounded-circle">
          <span class="username">${post.user}</span>
        </div>
        <img src="${post.image}" alt="Post">
        <div class="post-details">
          <p class="post-caption">${post.caption}</p>
          <div class="post-categories">
            ${categoryTags}
          </div>
          <div class="post-interactions">
            <div class="interaction-item like-btn" data-toggle="modal">
              <i class="fas fa-heart"></i> 
              <span>${post.likes}</span>
            </div>
            <div class="interaction-item dislike-btn">
              <i class="fas fa-thumbs-down"></i> 
              <span>${post.dislikes}</span>
            </div>
            <div class="interaction-item comment-btn">
              <i class="fas fa-comment"></i> 
              <span>${post.comments.length}</span>
            </div>
          </div>
          <div class="likes-section">
            <span class="toggle-likes-details">Liked by ${post.likedBy.slice(0, 3).join(', ')} 
              ${post.likedBy.length > 3 ? `and ${post.likedBy.length - 3} others` : ''}
            </span>
            <div class="likes-list hidden">
              ${this.renderLikesList(post.likedBy)}
            </div>
          </div>
        </div>
        <div class="post-comments-section">
          <div class="comments-toggle">
            View Comments (${post.comments.length})
          </div>
          <div class="comments-container hidden">
            ${this.renderComments(post.comments)}
          </div>
          <div class="comment-input-container">
            <input type="text" class="form-control comment-input" placeholder="Add a comment...">
            <button class="btn btn-secondary send-comment">Send</button>
          </div>
        </div>
      `;
  
      return postElement;
    }
  
    renderLikesList(likedBy) {
      return likedBy.map(user => `
        <div class="user-like">
          <span>${user}</span>
        </div>
      `).join('');
    }
  
    renderComments(comments) {
      return comments.map(comment => `
        <div class="comment">
          <strong>${comment.user}:</strong> ${comment.text}
        </div>
      `).join('');
    }
  
    addComment(post, postElement, commentText) {
      const newComment = {
        user: 'Current User',
        text: commentText
      };
      
      post.comments.push(newComment);
      
      // Update comments container
      const commentsContainer = postElement.querySelector('.comments-container');
      const commentToggle = postElement.querySelector('.comments-toggle');
      
      commentsContainer.innerHTML = this.renderComments(post.comments);
      
      // Update comment count
      const commentBtn = postElement.querySelector('.comment-btn span');
      const commentToggleText = postElement.querySelector('.comments-toggle');
      
      commentBtn.textContent = post.comments.length;
      commentToggleText.textContent = `View Comments (${post.comments.length})`;
      
      // Do not hide the comments section after adding a comment
      commentsContainer.classList.remove('hidden');
    }
  
    createPost(content, selectedCategories) {
      const newPost = {
        id: Date.now(),
        user: 'Current User',
        userAvatar: 'https://randomuser.me/api/portraits/women/44.jpg',
        image: 'https://via.placeholder.com/300x250?text=New+Post',
        caption: content,
        categories: selectedCategories,
        likes: 0,
        dislikes: 0,
        comments: [],
        likedBy: [],
        dislikedBy: []
      };
  
      this.posts.unshift(newPost);
      const postElement = this.createPostElement(newPost);
      this.domElements.postsContainer.prepend(postElement);
    }
  
    toggleComments(postElement) {
      const commentsContainer = postElement.querySelector('.comments-container');
      commentsContainer.classList.toggle('hidden');
    }
  
    handleCommentSubmission(postElement) {
      const commentInput = postElement.querySelector('.comment-input');
      const post = this.findPostFromElement(postElement);
      const commentText = commentInput.value.trim();
  
      if (commentText) {
        this.addComment(post, postElement, commentText);
      }
    }
  
    toggleLike(postElement) {
      const post = this.findPostFromElement(postElement);
      post.likes++;
      this.updateInteractionUI(postElement, 'like', post.likes);
    }
  
    toggleDislike(postElement) {
      const post = this.findPostFromElement(postElement);
      post.dislikes++;
      this.updateInteractionUI(postElement, 'dislike', post.dislikes);
    }
  
    findPostFromElement(postElement) {
      return this.posts.find(p => p.id == postElement.dataset.postId);
    }
  
    updateInteractionUI(postElement, type, count) {
      const interactionBtn = postElement.querySelector(`.${type}-btn span`);
      interactionBtn.textContent = count;
    }
  
    toggleLikesDetails(postElement) {
      const likesList = postElement.querySelector('.likes-list');
      likesList.classList.toggle('hidden');
    }
  
    filterPosts(filterType, selectedCategories = []) {
      // Clear current posts
      this.domElements.postsContainer.innerHTML = '';
  
      // Filter logic
      const filteredPosts = this.posts.filter(post => {
        switch(filterType) {
          case 'created':
            return post.user === 'Current User';
          case 'liked':
            return post.likedBy.includes('Current User');
          case 'categories':
            return selectedCategories.length === 0 || 
                   selectedCategories.some(category => post.categories.includes(category));
          default:
            return true;
        }
      });
  
      // Render filtered posts
      filteredPosts.forEach(post => {
        const postElement = this.createPostElement(post);
        this.domElements.postsContainer.appendChild(postElement);
      });
    }
  }
  
  document.addEventListener('DOMContentLoaded', () => {
    new SocialApp();
  });
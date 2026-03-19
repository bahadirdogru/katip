export interface DiffItem {
  type: 'equal' | 'insert' | 'delete';
  text: string;
}

export interface Review {
  id: string;
  paragraphId: string;
  summary: string;
  original: string;
  improved: string;
  diffs: DiffItem[];
  status: 'pending' | 'accepted' | 'rejected';
}

class ReviewStore {
  reviews: Review[] = $state([]);

  addReview(review: Omit<Review, 'id' | 'status'>) {
    const id = `review-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;
    this.reviews.push({
      ...review,
      id,
      status: 'pending',
    });
    return id;
  }

  acceptReview(id: string) {
    const review = this.reviews.find(r => r.id === id);
    if (review) {
      review.status = 'accepted';
    }
  }

  rejectReview(id: string) {
    const review = this.reviews.find(r => r.id === id);
    if (review) {
      review.status = 'rejected';
    }
  }

  removeReview(id: string) {
    this.reviews = this.reviews.filter(r => r.id !== id);
  }

  get pendingReviews(): Review[] {
    return this.reviews.filter(r => r.status === 'pending');
  }

  clear() {
    this.reviews = [];
  }
}

export const reviewStore = new ReviewStore();

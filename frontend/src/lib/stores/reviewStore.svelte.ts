export interface DiffItem {
  type: 'equal' | 'insert' | 'delete';
  text: string;
  rejected?: boolean;
  accepted?: boolean;
}

export type ImprovementMode = 'fix' | 'shorten' | 'flow' | 'formal';
export type ReviewKind = 'improvement' | 'consistency' | 'style';
export type ChangeType = 'spelling' | 'grammar' | 'style' | 'consistency';

export interface Review {
  id: string;
  paragraphId: string;
  paragraphPos: number;
  selectionFrom?: number;
  selectionTo?: number;
  summary: string;
  original: string;
  improved: string;
  diffs: DiffItem[];
  status: 'pending' | 'accepted' | 'rejected' | 'partial';
  mode?: ImprovementMode;
  kind?: ReviewKind;
  changeType?: ChangeType;
  ruleName?: string;
}

class ReviewStore {
  reviews: Review[] = $state([]);
  activeReviewId: string | null = $state(null);

  addReview(review: Omit<Review, 'id' | 'status'>) {
    const id = `review-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;
    this.reviews.push({
      ...review,
      id,
      status: 'pending',
      kind: review.kind ?? 'improvement',
    });
    this.activeReviewId = id;
    return id;
  }

  acceptReview(id: string) {
    const review = this.reviews.find((r) => r.id === id);
    if (review) {
      review.status = 'accepted';
      review.diffs.forEach((d) => {
        if (d.type !== 'equal') d.accepted = true;
      });
    }
  }

  rejectReview(id: string) {
    const review = this.reviews.find((r) => r.id === id);
    if (review) {
      review.status = 'rejected';
      review.diffs.forEach((d) => {
        if (d.type !== 'equal') d.rejected = true;
      });
    }
  }

  acceptDiffItem(reviewId: string, diffIndex: number) {
    const review = this.reviews.find((r) => r.id === reviewId);
    if (!review) return;
    const diff = review.diffs[diffIndex];
    if (diff && diff.type !== 'equal') {
      diff.accepted = true;
      diff.rejected = false;
    }
    this.updatePartialStatus(review);
  }

  rejectDiffItem(reviewId: string, diffIndex: number) {
    const review = this.reviews.find((r) => r.id === reviewId);
    if (!review) return;
    const diff = review.diffs[diffIndex];
    if (diff && diff.type !== 'equal') {
      diff.rejected = true;
      diff.accepted = false;
    }
    this.updatePartialStatus(review);
  }

  private updatePartialStatus(review: Review) {
    const changes = review.diffs.filter((d) => d.type !== 'equal');
    const allAccepted = changes.every((d) => d.accepted);
    const allRejected = changes.every((d) => d.rejected);
    if (allAccepted) review.status = 'accepted';
    else if (allRejected) review.status = 'rejected';
    else review.status = 'partial';
  }

  buildPartialImproved(review: Review): string {
    let result = '';
    for (const d of review.diffs) {
      if (d.type === 'equal') {
        result += d.text;
      } else if (d.type === 'delete') {
        if (d.rejected) result += d.text;
      } else if (d.type === 'insert') {
        if (d.accepted) result += d.text;
      }
    }
    return result;
  }

  acceptAllPending() {
    for (const r of this.pendingReviews) {
      this.acceptReview(r.id);
    }
  }

  rejectAllPending() {
    for (const r of this.pendingReviews) {
      this.rejectReview(r.id);
    }
  }

  clearCompleted() {
    this.reviews = this.reviews.filter((r) => r.status === 'pending' || r.status === 'partial');
  }

  removeReview(id: string) {
    this.reviews = this.reviews.filter((r) => r.id !== id);
    if (this.activeReviewId === id) {
      this.activeReviewId = this.pendingReviews[0]?.id ?? null;
    }
  }

  setActiveReview(id: string | null) {
    this.activeReviewId = id;
  }

  get pendingReviews(): Review[] {
    return this.reviews.filter((r) => r.status === 'pending' || r.status === 'partial');
  }

  get nextPending(): Review | undefined {
    const pending = this.pendingReviews;
    if (!this.activeReviewId) return pending[0];
    const idx = pending.findIndex((r) => r.id === this.activeReviewId);
    return pending[idx + 1] ?? pending[0];
  }

  get prevPending(): Review | undefined {
    const pending = this.pendingReviews;
    if (!this.activeReviewId) return pending[pending.length - 1];
    const idx = pending.findIndex((r) => r.id === this.activeReviewId);
    return pending[idx - 1] ?? pending[pending.length - 1];
  }

  clear() {
    this.reviews = [];
    this.activeReviewId = null;
  }
}

export const reviewStore = new ReviewStore();

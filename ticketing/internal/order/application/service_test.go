package application

import (
	"context"
	"errors"
	"testing"
)

func TestPaymentCallback_InvalidSignatureRejected(t *testing.T) {
	t.Parallel()

	svc := &Service{
		cfg: Config{
			PaymentSignKey: "unit-test-key",
		},
	}

	_, err := svc.PaymentCallback(context.Background(), PaymentCallbackInput{
		OrderID:       "order-1",
		ProviderTxnID: "txn-1",
		Status:        "SUCCESS",
		Signature:     "bad-signature",
	})
	if !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("expected ErrInvalidSignature, got: %v", err)
	}
}

func TestPaymentCallback_StatusWhitelist(t *testing.T) {
	t.Parallel()

	svc := &Service{
		cfg: Config{
			PaymentSignKey: "unit-test-key",
		},
	}
	signature := signHMACSHA256("unit-test-key", "order-2|txn-2|FAILED")

	_, err := svc.PaymentCallback(context.Background(), PaymentCallbackInput{
		OrderID:       "order-2",
		ProviderTxnID: "txn-2",
		Status:        "FAILED",
		Signature:     signature,
	})
	if !errors.Is(err, ErrInvalidPaymentStatus) {
		t.Fatalf("expected ErrInvalidPaymentStatus, got: %v", err)
	}
}


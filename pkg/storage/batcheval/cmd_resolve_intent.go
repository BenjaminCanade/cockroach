// Copyright 2014 The Cockroach Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

package batcheval

import (
	"golang.org/x/net/context"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage/batcheval/result"
	"github.com/cockroachdb/cockroach/pkg/storage/engine"
)

// ResolveIntent resolves a write intent from the specified key
// according to the status of the transaction which created it.
func ResolveIntent(
	ctx context.Context, batch engine.ReadWriter, cArgs CommandArgs, resp roachpb.Response,
) (result.Result, error) {
	args := cArgs.Args.(*roachpb.ResolveIntentRequest)
	h := cArgs.Header
	ms := cArgs.Stats

	if h.Txn != nil {
		return result.Result{}, ErrTransactionUnsupported
	}

	intent := roachpb.Intent{
		Span:   args.Span,
		Txn:    args.IntentTxn,
		Status: args.Status,
	}
	if err := engine.MVCCResolveWriteIntent(ctx, batch, ms, intent); err != nil {
		return result.Result{}, err
	}
	if WriteAbortSpanOnResolve(args.Status) {
		return result.Result{}, SetAbortSpan(ctx, cArgs.EvalCtx, batch, ms, args.IntentTxn, args.Poison)
	}
	return result.Result{}, nil
}

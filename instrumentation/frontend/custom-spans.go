func (fe *frontendServer) viewCartHandler(w http.ResponseWriter, r *http.Request) {
  tracer := otel.Tracer("frontend")

  // Custom span 1: validate-cart-contents
  ctx, span := tracer.Start(r.Context(), "validate-cart-contents")
  defer span.End()

  userID := sessionID(r)
  span.SetAttributes(
    attribute.String("user.id", userID),
    attribute.String("page.route", "/cart"),
  )

  cart, err := fe.getCart(ctx, userID)
  if err != nil {
    span.RecordError(err)
    span.SetStatus(codes.Error, err.Error())
    renderHTTPError(w, r, err, http.StatusInternalServerError)
    return
  }

  // Custom span 2: calculate-shipping-cost (child of span 1)
  _, shippingSpan := tracer.Start(ctx, "calculate-shipping-cost")
  shippingCost, _ := fe.getShippingQuote(ctx, cart.Items, "NG")
  shippingSpan.SetAttributes(
    attribute.Float64("shipping.cost_usd", float64(shippingCost.GetUnits())),
    attribute.Int("cart.item_count", len(cart.Items)),
  )
  shippingSpan.End()

  // Custom metric: cart.views counter
  meter := otel.Meter("frontend")
  counter, _ := meter.Int64Counter("frontend.cart.views",
    metric.WithDescription("Number of cart page views"),
    metric.WithUnit("1"),
  )
  counter.Add(ctx, 1, metric.WithAttributes(
    attribute.String("user.id", userID),
  ))
}
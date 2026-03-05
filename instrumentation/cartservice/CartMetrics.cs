using System.Diagnostics;
using System.Diagnostics.Metrics;

public class CartInstrumentation : IDisposable
{
    private static readonly string ServiceName = "cartservice";
    private static readonly string ServiceVersion = "1.0.0";

    public static readonly ActivitySource ActivitySource =
        new ActivitySource(ServiceName, ServiceVersion);

    private static readonly Meter Meter = new(ServiceName, ServiceVersion);

    // Custom metric 1: items added counter
    public static readonly Counter<long> ItemsAdded =
        Meter.CreateCounter<long>("cart.items.added",
            description: "Total items added to carts");

    // Custom metric 2: cart size at checkout histogram
    public static readonly Histogram<double> CartSize =
        Meter.CreateHistogram<double>("cart.size",
            unit: "items",
            description: "Number of items in cart at checkout time");

    // Custom span 1: validate-cart-contents
    public static Activity? StartValidateCartSpan(string userId)
    {
        var activity = ActivitySource.StartActivity("validate-cart-contents");
        activity?.SetTag("user.id", userId);
        activity?.SetTag("cart.operation", "validate");
        return activity;
    }

    // Custom span 2: calculate-cart-total
    public static Activity? StartCartTotalSpan(int itemCount, double total)
    {
        var activity = ActivitySource.StartActivity("calculate-cart-total");
        activity?.SetTag("cart.item_count", itemCount);
        activity?.SetTag("cart.total_usd", total);
        return activity;
    }

    public static void RecordItemAdded(string userId, string productId) =>
        ItemsAdded.Add(1,
            new KeyValuePair<string, object?>("user.id", userId),
            new KeyValuePair<string, object?>("product.id", productId));

    public static void RecordCartCheckoutSize(int size, string userId) =>
        CartSize.Record(size, new KeyValuePair<string, object?>("user.id", userId));

    public void Dispose()
    {
        ActivitySource.Dispose();
        Meter.Dispose();
    }
}
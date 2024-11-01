using System.Text.Json.Nodes;
using System.Text.Json.Serialization;
using TUI.Engine;

namespace TUI.Providers.Dependencies;

public class PackageJson
{
    [JsonPropertyName("dependencies")]
    public JsonObject? Dependencies { get; set; }

    [JsonPropertyName("devDependencies")]
    public JsonObject? DevDependencies { get; set; }

    [JsonPropertyName("engines")]
    public JsonObject? Engines { get; set; }
    
    public Version? ParseVersion(string? dependencyName)
    {
        if (dependencyName == null) return null;
        
        JsonNode? version = null;
        
        var lowerDependencyName = dependencyName.ToLower();
        Dependencies?.TryGetPropertyValue(lowerDependencyName, out version);
        
        if (version == null) Engines?.TryGetPropertyValue(lowerDependencyName, out version);
        
        if (version == null) DevDependencies?.TryGetPropertyValue(lowerDependencyName, out version);
        
        return version?.GetValue<string>().ToVersion();
    }
}
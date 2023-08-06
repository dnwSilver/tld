using System.Text.Json.Nodes;
using System.Text.Json.Serialization;


namespace TUI.Domain;

public class Package
{
    [JsonPropertyName("dependencies")]
    public JsonObject Dependencies { get; set; }
}
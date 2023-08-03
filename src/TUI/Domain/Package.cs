using System.Text.Json.Serialization;


namespace TUI.Domain;

public class Package
{
    [JsonPropertyName("dependencies")]
    public Dependencies Dependencies { get; set; }
}

public class Dependencies
{
    [JsonPropertyName("react")]
    public string React { get; set; }

    [JsonPropertyName("typesciprt")]
    public string TypeScript { get; set; }

    public string GetVersion(Dependency dependency) => dependency.Name.ToLower() switch
    {
        "react"     => React,
        "typescipt" => TypeScript,
        _           => "-"
    };
}